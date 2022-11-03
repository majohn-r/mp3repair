package output

import "testing"

func TestNewNilBus(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "simple test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewNilBus()
			o.WriteCanonicalConsole("%s %d %t", "foo", 42, true)
			o.WriteConsole("%s %d %t", "foo", 42, true)
			o.WriteCanonicalError("%s %d %t", "foo", 42, true)
			o.WriteError("%s %d %t", "foo", 42, true)
			o.LogWriter().Error("error message", map[string]any{"field1": "value"})
		})
	}
}

func TestNilWriter_Write(t *testing.T) {
	fnName := "NilWriter.Write()"
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		nw   NilWriter
		args
		wantN   int
		wantErr bool
	}{
		{
			name: "a few bytes",
			nw:   NilWriter{},
			args: args{
				p: []byte{0, 1, 2},
			},
			wantN:   3,
			wantErr: false,
		},
		{
			name: "nil",
			nw:   NilWriter{},
			args: args{
				p: nil,
			},
			wantN:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotN, err := tt.nw.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("%s = %v, want %v", fnName, gotN, tt.wantN)
			}
		})
	}
}
