package files

import (
	"path/filepath"
	"testing"
)

func Test_isIllegalRuneForFileNames(t *testing.T) {
	fnName := "isIllegalRuneForFileNames()"
	type args struct {
		r rune
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "0", args: args{r: 0}, want: true},
		{name: "1", args: args{r: 1}, want: true},
		{name: "2", args: args{r: 2}, want: true},
		{name: "3", args: args{r: 3}, want: true},
		{name: "4", args: args{r: 4}, want: true},
		{name: "5", args: args{r: 5}, want: true},
		{name: "6", args: args{r: 6}, want: true},
		{name: "7", args: args{r: 7}, want: true},
		{name: "8", args: args{r: 8}, want: true},
		{name: "9", args: args{r: 9}, want: true},
		{name: "10", args: args{r: 10}, want: true},
		{name: "11", args: args{r: 11}, want: true},
		{name: "12", args: args{r: 12}, want: true},
		{name: "13", args: args{r: 13}, want: true},
		{name: "14", args: args{r: 14}, want: true},
		{name: "15", args: args{r: 15}, want: true},
		{name: "16", args: args{r: 16}, want: true},
		{name: "17", args: args{r: 17}, want: true},
		{name: "18", args: args{r: 18}, want: true},
		{name: "19", args: args{r: 19}, want: true},
		{name: "20", args: args{r: 20}, want: true},
		{name: "21", args: args{r: 21}, want: true},
		{name: "22", args: args{r: 22}, want: true},
		{name: "23", args: args{r: 23}, want: true},
		{name: "24", args: args{r: 24}, want: true},
		{name: "25", args: args{r: 25}, want: true},
		{name: "26", args: args{r: 26}, want: true},
		{name: "27", args: args{r: 27}, want: true},
		{name: "28", args: args{r: 28}, want: true},
		{name: "29", args: args{r: 29}, want: true},
		{name: "30", args: args{r: 30}, want: true},
		{name: "31", args: args{r: 31}, want: true},
		{name: "<", args: args{r: '<'}, want: true},
		{name: ">", args: args{r: '>'}, want: true},
		{name: ":", args: args{r: ':'}, want: true},
		{name: "\"", args: args{r: '"'}, want: true},
		{name: "/", args: args{r: '/'}, want: true},
		{name: "\\", args: args{r: '\\'}, want: true},
		{name: "|", args: args{r: '|'}, want: true},
		{name: "?", args: args{r: '?'}, want: true},
		{name: "*", args: args{r: '*'}, want: true},
		{name: "!", args: args{r: '!'}, want: false},
		{name: "#", args: args{r: '#'}, want: false},
		{name: "$", args: args{r: '$'}, want: false},
		{name: "&", args: args{r: '&'}, want: false},
		{name: "'", args: args{r: '\''}, want: false},
		{name: "(", args: args{r: '('}, want: false},
		{name: ")", args: args{r: ')'}, want: false},
		{name: "+", args: args{r: '+'}, want: false},
		{name: ",", args: args{r: ','}, want: false},
		{name: "-", args: args{r: '-'}, want: false},
		{name: ".", args: args{r: '.'}, want: false},
		{name: "0", args: args{r: '0'}, want: false},
		{name: "1", args: args{r: '1'}, want: false},
		{name: "2", args: args{r: '2'}, want: false},
		{name: "3", args: args{r: '3'}, want: false},
		{name: "4", args: args{r: '4'}, want: false},
		{name: "5", args: args{r: '5'}, want: false},
		{name: "6", args: args{r: '6'}, want: false},
		{name: "7", args: args{r: '7'}, want: false},
		{name: "8", args: args{r: '8'}, want: false},
		{name: "9", args: args{r: '9'}, want: false},
		{name: ";", args: args{r: ';'}, want: false},
		{name: "A", args: args{r: 'A'}, want: false},
		{name: "B", args: args{r: 'B'}, want: false},
		{name: "C", args: args{r: 'C'}, want: false},
		{name: "D", args: args{r: 'D'}, want: false},
		{name: "E", args: args{r: 'E'}, want: false},
		{name: "F", args: args{r: 'F'}, want: false},
		{name: "G", args: args{r: 'G'}, want: false},
		{name: "H", args: args{r: 'H'}, want: false},
		{name: "I", args: args{r: 'I'}, want: false},
		{name: "J", args: args{r: 'J'}, want: false},
		{name: "K", args: args{r: 'K'}, want: false},
		{name: "L", args: args{r: 'L'}, want: false},
		{name: "M", args: args{r: 'M'}, want: false},
		{name: "N", args: args{r: 'N'}, want: false},
		{name: "O", args: args{r: 'O'}, want: false},
		{name: "P", args: args{r: 'P'}, want: false},
		{name: "Q", args: args{r: 'Q'}, want: false},
		{name: "R", args: args{r: 'R'}, want: false},
		{name: "S", args: args{r: 'S'}, want: false},
		{name: "T", args: args{r: 'T'}, want: false},
		{name: "U", args: args{r: 'U'}, want: false},
		{name: "V", args: args{r: 'V'}, want: false},
		{name: "W", args: args{r: 'W'}, want: false},
		{name: "X", args: args{r: 'X'}, want: false},
		{name: "Y", args: args{r: 'Y'}, want: false},
		{name: "Z", args: args{r: 'Z'}, want: false},
		{name: "[", args: args{r: '['}, want: false},
		{name: "]", args: args{r: ']'}, want: false},
		{name: "_", args: args{r: '_'}, want: false},
		{name: "a", args: args{r: 'a'}, want: false},
		{name: "b", args: args{r: 'b'}, want: false},
		{name: "c", args: args{r: 'c'}, want: false},
		{name: "d", args: args{r: 'd'}, want: false},
		{name: "e", args: args{r: 'e'}, want: false},
		{name: "f", args: args{r: 'f'}, want: false},
		{name: "g", args: args{r: 'g'}, want: false},
		{name: "h", args: args{r: 'h'}, want: false},
		{name: "i", args: args{r: 'i'}, want: false},
		{name: "j", args: args{r: 'j'}, want: false},
		{name: "k", args: args{r: 'k'}, want: false},
		{name: "l", args: args{r: 'l'}, want: false},
		{name: "m", args: args{r: 'm'}, want: false},
		{name: "n", args: args{r: 'n'}, want: false},
		{name: "o", args: args{r: 'o'}, want: false},
		{name: "p", args: args{r: 'p'}, want: false},
		{name: "q", args: args{r: 'q'}, want: false},
		{name: "r", args: args{r: 'r'}, want: false},
		{name: "s", args: args{r: 's'}, want: false},
		{name: "space", args: args{r: ' '}, want: false},
		{name: "t", args: args{r: 't'}, want: false},
		{name: "u", args: args{r: 'u'}, want: false},
		{name: "v", args: args{r: 'v'}, want: false},
		{name: "w", args: args{r: 'w'}, want: false},
		{name: "x", args: args{r: 'x'}, want: false},
		{name: "y", args: args{r: 'y'}, want: false},
		{name: "z", args: args{r: 'z'}, want: false},
		{name: "Á", args: args{r: 'Á'}, want: false},
		{name: "È", args: args{r: 'È'}, want: false},
		{name: "É", args: args{r: 'É'}, want: false},
		{name: "Ô", args: args{r: 'Ô'}, want: false},
		{name: "à", args: args{r: 'à'}, want: false},
		{name: "á", args: args{r: 'á'}, want: false},
		{name: "ã", args: args{r: 'ã'}, want: false},
		{name: "ä", args: args{r: 'ä'}, want: false},
		{name: "å", args: args{r: 'å'}, want: false},
		{name: "ç", args: args{r: 'ç'}, want: false},
		{name: "è", args: args{r: 'è'}, want: false},
		{name: "é", args: args{r: 'é'}, want: false},
		{name: "ê", args: args{r: 'ê'}, want: false},
		{name: "ë", args: args{r: 'ë'}, want: false},
		{name: "í", args: args{r: 'í'}, want: false},
		{name: "î", args: args{r: 'î'}, want: false},
		{name: "ï", args: args{r: 'ï'}, want: false},
		{name: "ñ", args: args{r: 'ñ'}, want: false},
		{name: "ò", args: args{r: 'ò'}, want: false},
		{name: "ó", args: args{r: 'ó'}, want: false},
		{name: "ô", args: args{r: 'ô'}, want: false},
		{name: "ö", args: args{r: 'ö'}, want: false},
		{name: "ø", args: args{r: 'ø'}, want: false},
		{name: "ù", args: args{r: 'ù'}, want: false},
		{name: "ú", args: args{r: 'ú'}, want: false},
		{name: "ü", args: args{r: 'ü'}, want: false},
		{name: "ř", args: args{r: 'ř'}, want: false},
		{name: "…", args: args{r: '…'}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIllegalRuneForFileNames(tt.args.r); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestCreateBackupPath(t *testing.T) {
	fnName := "CreateBackupPath()"
	type args struct {
		topDir string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple test",
			args: args{topDir: "top-level-directory"},
			want: filepath.Join("top-level-directory", backupDirName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateBackupPath(tt.args.topDir); got != tt.want {
				t.Errorf("%s = %q, want %q", fnName, got, tt.want)
			}
		})
	}
}
