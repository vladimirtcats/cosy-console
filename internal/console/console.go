// Package console runs an interactive interpreter session over the
// application composition root: the whole container is exported as a single
// live value App, plus Ctx for calls that require a context.Context.
//
// Domain packages registered by symbols.Symbols and stdlib.Symbols are
// available without import statements: ImportUsed binds every registered
// package by the last segment of its key.
package console

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"

	"cosy-console/internal/app"
	"cosy-console/internal/console/symbols"
)

// consolePath is the pseudo import path the App and Ctx values are registered
// under; the session dot-imports it to expose both names without a prefix.
const consolePath = "console/console"

type Options struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Eval   string
}

// Run evaluates a single statement from Options.Eval, or runs the
// interactive REPL until stdin reaches EOF (Ctrl+D).
//
// SIGINT (Ctrl+C) is trapped by the REPL itself: it cancels the running
// statement and returns to the prompt. The passed ctx must therefore NOT be
// cancelled by SIGINT — otherwise the exported Ctx dies with the first
// Ctrl+C and every following App call fails with context.Canceled.
func Run(ctx context.Context, container *app.Container, opts Options) error {
	i, err := newInterpreter(ctx, container, opts)
	if err != nil {
		return err
	}

	if opts.Eval != "" {
		return runEval(ctx, i, opts.Eval, opts.Stdout)
	}

	return runREPL(ctx, i)
}

func newInterpreter(ctx context.Context, container *app.Container, opts Options) (*interp.Interpreter, error) {
	i := interp.New(interp.Options{
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
	})

	// Three symbol sources, three Use calls: stdlib (fmt, time, ...), the
	// domain packages (after the Exports() re-keying), and the live App/Ctx
	// values. Same wrapping, different messages — a failed registration
	// names the guilty table. Exports() is a separate step on purpose: it
	// fails on its own (duplicate session names), which is a different
	// error from "the interpreter rejected the table".
	if err := i.Use(stdlib.Symbols); err != nil {
		return nil, fmt.Errorf("use stdlib symbols: %w", err)
	}

	symbolsExports, err := symbols.Exports()
	if err != nil {
		return nil, fmt.Errorf("build console symbols: %w", err)
	}

	if err := i.Use(symbolsExports); err != nil {
		return nil, fmt.Errorf("use app symbols: %w", err)
	}

	if err := i.Use(exports(ctx, container)); err != nil {
		return nil, fmt.Errorf("use console exports: %w", err)
	}

	// Bind every registered package by the last segment of its key, the way
	// yaegi's own CLI REPL does. Strictly before the first Eval: the doc
	// comment forbids calling ImportUsed after it, as it may rename packages.
	i.ImportUsed()

	// App and Ctx are values, not a package: ImportUsed cannot expose them,
	// so the session dot-imports the pseudo package once, by hand.
	if _, err := i.Eval(`import . "console"`); err != nil {
		return nil, fmt.Errorf("import console: %w", err)
	}

	return i, nil
}

// runREPL runs the interactive REPL until stdin reaches EOF or ctx is done.
// The error returned by i.REPL on EOF is the error of the last evaluated
// statement — already printed to stderr by the REPL — so it is discarded:
// a clean Ctrl+D exit must not fail the process.
func runREPL(ctx context.Context, i *interp.Interpreter) error {
	done := make(chan struct{})

	go func() {
		defer close(done)

		_, _ = i.REPL()
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func runEval(ctx context.Context, i *interp.Interpreter, code string, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}

	v, err := i.EvalWithContext(ctx, code)
	if err != nil {
		return fmt.Errorf("eval: %w", err)
	}

	if v.IsValid() {
		if _, err := fmt.Fprintln(out, ":", v); err != nil {
			return fmt.Errorf("print result: %w", err)
		}
	}

	return nil
}

func exports(ctx context.Context, container *app.Container) interp.Exports {
	return interp.Exports{
		consolePath: {
			"App": reflect.ValueOf(container),
			"Ctx": reflect.ValueOf(ctx),
		},
	}
}
