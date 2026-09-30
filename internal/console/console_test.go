package console_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosy-console/internal/app"
	"cosy-console/internal/console"
)

func newOptions(stdout *bytes.Buffer) console.Options {
	return console.Options{
		Stdin:  bytes.NewReader(nil),
		Stdout: stdout,
		Stderr: stdout,
	}
}

func TestRun_EvalMode(t *testing.T) {
	tests := []struct {
		name    string
		eval    string
		wantOut string
		wantErr bool
	}{
		{
			name:    "expression result is printed",
			eval:    "1 + 1",
			wantOut: ": 2\n",
		},
		{
			name:    "stdlib fmt is available without import",
			eval:    `fmt.Sprintf("%s-%d", "a", 1)`,
			wantOut: ": a-1\n",
		},
		{
			name:    "domain params type is available without import",
			eval:    "orders.UpdateAttrs{}.Note",
			wantOut: ": <nil>\n",
		},
		{
			name:    "domain models are bound under a unique package name",
			eval:    "orders_models.Order{Note: \"n\"}.Note",
			wantOut: ": n\n",
		},
		{
			name:    "catalog filter is constructible",
			eval:    "catalog.ItemFilter{SKU: \"MUG\"}.SKU",
			wantOut: ": MUG\n",
		},
		{
			name:    "uuid package is available for method arguments",
			eval:    `uuid.MustParse("00000000-0000-0000-0000-000000000001").String()`,
			wantOut: ": 00000000-0000-0000-0000-000000000001\n",
		},
		{
			name:    "pagination options are callable",
			eval:    "pagination.NewSettings(pagination.WithPage(3), pagination.WithPageSize(5)).PageSize",
			wantOut: ": 5\n",
		},
		{
			name:    "invalid code returns error",
			eval:    "this is not go",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := newOptions(&stdout)
			opts.Eval = tt.eval

			err := console.Run(context.Background(), &app.Container{}, opts)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantOut, stdout.String())
		})
	}
}

func TestRun_EvalModeExports(t *testing.T) {
	tests := []struct {
		name    string
		eval    string
		wantOut string
	}{
		{
			name:    "App export resolves to the container value",
			eval:    "fmt.Sprintf(\"%T\", App)",
			wantOut: ": *app.Container\n",
		},
		{
			name:    "Ctx export resolves to a context",
			eval:    "fmt.Sprintf(\"%T\", Ctx)",
			wantOut: ": context.backgroundCtx\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := newOptions(&stdout)
			opts.Eval = tt.eval

			err := console.Run(context.Background(), &app.Container{}, opts)
			require.NoError(t, err)
			assert.True(t, strings.HasSuffix(stdout.String(), tt.wantOut),
				"output %q must end with %q", stdout.String(), tt.wantOut)
		})
	}
}

func TestRun_REPL(t *testing.T) {
	t.Setenv("YAEGI_PROMPT", "1")

	tests := []struct {
		name    string
		input   string
		wantOut string
	}{
		{
			name:    "expressions are evaluated and results printed",
			input:   "1 + 1\n",
			wantOut: ": 2",
		},
		{
			name:    "variables persist between lines",
			input:   "x := 21\nx * 2\n",
			wantOut: ": 42",
		},
		{
			name:    "App is reachable from the REPL",
			input:   "App != nil\n",
			wantOut: ": true",
		},
		{
			name:    "multi-line function literal is buffered",
			input:   "double := func(x int) int {\nreturn x * 2\n}\ndouble(21)\n",
			wantOut: ": 42",
		},
		{
			// The stock REPL cannot buffer a multi-line closure used as a
			// CALL ARGUMENT (intermediate parses fail with "missing ','
			// before newline in argument list" and reset the buffer). The
			// working pattern: define the closure in an assignment — any
			// body length is buffered fine — then call it on one line.
			name: "transaction pattern: multi-line closure assignment plus one-line call",
			input: "id, _ := uuid.Parse(\"00000000-0000-0000-0000-000000000000\")\n" +
				"probe := func(ctx context.Context) error {\n" +
				"if ctx == nil {\n" +
				"return errors.New(\"no ctx\")\n" +
				"}\n" +
				"return nil\n" +
				"}\n" +
				"err := probe(Ctx)\n" +
				"err\n",
			wantOut: ": <nil>",
		},
		{
			name:    "runtime error does not kill the session",
			input:   "var zero int\n1 / zero\n1 + 1\n",
			wantOut: ": 2",
		},
		{
			name:    "EOF exit after a failed statement is clean",
			input:   "NoSuchSymbol\n",
			wantOut: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := newOptions(&stdout)
			opts.Stdin = strings.NewReader(tt.input)

			err := console.Run(context.Background(), &app.Container{}, opts)

			require.NoError(t, err)
			assert.Contains(t, stdout.String(), tt.wantOut)
		})
	}
}
