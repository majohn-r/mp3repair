package output

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNewRecorder(t *testing.T) {
	fnName := "NewRecorder()"
	tests := []struct {
		name            string
		canonicalWrites bool
		consoleFmt      string
		consoleArgs     []any
		errorFmt        string
		errorArgs       []any
		logMessage      string
		logArgs         map[string]any
		WantedRecording
	}{
		{
			name:        "non-canonical test",
			consoleFmt:  "%s %d %t",
			consoleArgs: []any{"hello", 42, true},
			errorFmt:    "%d %t %s",
			errorArgs:   []any{24, false, "bye"},
			logMessage:  "hello!",
			logArgs:     map[string]any{"field": "value"},
			WantedRecording: WantedRecording{
				Console: "hello 42 true",
				Error:   "24 false bye",
				Log:     "level='error' field='value' msg='hello!'\n",
			},
		},
		{
			name:            "canonical test",
			canonicalWrites: true,
			consoleFmt:      "%s %d %t",
			consoleArgs:     []any{"hello", 42, true},
			errorFmt:        "%d %t %s",
			errorArgs:       []any{24, false, "bye"},
			logMessage:      "hello!",
			logArgs:         map[string]any{"field": "value"},
			WantedRecording: WantedRecording{
				Console: "Hello 42 true.\n",
				Error:   "24 false bye.\n",
				Log:     "level='error' field='value' msg='hello!'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewRecorder()
			var i any = o
			if _, ok := i.(Bus); !ok {
				t.Errorf("%s: Recorder does not implement Bus", fnName)
			}
			if o.ConsoleWriter() == nil {
				t.Errorf("%s: console writer is nil", fnName)
			}
			if o.ErrorWriter() == nil {
				t.Errorf("%s: error writer is nil", fnName)
			}
			if o.LogWriter() == nil {
				t.Errorf("%s: log writer is nil", fnName)
			}
			if tt.canonicalWrites {
				o.WriteCanonicalConsole(tt.consoleFmt, tt.consoleArgs...)
				o.WriteCanonicalError(tt.errorFmt, tt.errorArgs...)
			} else {
				o.WriteConsole(tt.consoleFmt, tt.consoleArgs...)
				o.WriteError(tt.errorFmt, tt.errorArgs...)
			}
			o.LogWriter().Error(tt.logMessage, tt.logArgs)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func TestRecorder_Verify(t *testing.T) {
	fnName := "Recorder.Verify()"
	type args struct {
		o *Recorder
		w WantedRecording
	}
	tests := []struct {
		name string
		args
		wantIssues []string
		wantOk     bool
	}{
		{name: "normal", args: args{o: NewRecorder(), w: WantedRecording{}}, wantOk: true},
		{
			name: "errors",
			args: args{
				o: NewRecorder(),
				w: WantedRecording{
					Console: "unexpected console output",
					Error:   "unexpected error output",
					Log:     "unexpected log output",
				},
			},
			wantIssues: []string{
				"console output = \"\", want \"unexpected console output\"",
				"error output = \"\", want \"unexpected error output\"",
				"log output = \"\", want \"unexpected log output\"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIssues, gotOk := tt.args.o.Verify(tt.args.w)
			if !reflect.DeepEqual(gotIssues, tt.wantIssues) {
				t.Errorf("%s gotIssues = %v, want %v", fnName, gotIssues, tt.wantIssues)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func TestRecorder_Log(t *testing.T) {
	fnName := "Recorder.Log()"
	type args struct {
		l    Level
		msg  string
		args map[string]any
	}
	tests := []struct {
		name string
		args
		Log   string
		Error string
	}{
		{
			name: "trace",
			args: args{
				l:    Trace,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='trace' f='v' msg='hello'\n",
		},
		{
			name: "debug",
			args: args{
				l:    Debug,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='debug' f='v' msg='hello'\n",
		},
		{
			name: "info",
			args: args{
				l:    Info,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='info' f='v' msg='hello'\n",
		},
		{
			name: "warning",
			args: args{
				l:    Warning,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='warning' f='v' msg='hello'\n",
		},
		{
			name: "error",
			args: args{
				l:    Error,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='error' f='v' msg='hello'\n",
		},
		{
			name: "panic",
			args: args{
				l:    Panic,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='panic' f='v' msg='hello'\n",
		},
		{
			name: "fatal",
			args: args{
				l:    Fatal,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='fatal' f='v' msg='hello'\n",
		},
		{
			name: "illegal",
			args: args{
				l:    Trace + 1,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Error: "Programming error: call to Recorder.Log() with invalid level value 7; message: 'hello', args: 'map[f:v].\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRecorder()
			r.Log(tt.args.l, tt.args.msg, tt.args.args)
			if got := r.LogOutput(); got != tt.Log {
				t.Errorf("%s got log %q want %q", fnName, got, tt.Log)
			}
			if got := r.ErrorOutput(); got != tt.Error {
				t.Errorf("%s got error %q want %q", fnName, got, tt.Error)
			}
		})
	}
}

func TestRecordingLogger_Trace(t *testing.T) {
	fnName := "RecordingLogger.Trace()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='trace' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Trace(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Debug(t *testing.T) {
	fnName := "RecordingLogger.Debug()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='debug' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Debug(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Info(t *testing.T) {
	fnName := "RecordingLogger.Info()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='info' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Info(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Warning(t *testing.T) {
	fnName := "RecordingLogger.Warning()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='warning' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Warning(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Error(t *testing.T) {
	fnName := "RecordingLogger.Error()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='error' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Error(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Panic(t *testing.T) {
	fnName := "RecordingLogger.Panic()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='panic' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Panic(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Fatal(t *testing.T) {
	fnName := "RecordingLogger.Fatal()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='fatal' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Fatal(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}
