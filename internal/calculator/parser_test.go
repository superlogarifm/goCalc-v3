package calculator

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "простое выражение",
			expr:    "2+2",
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "выражение с пробелами",
			expr:    "2 + 2",
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "сложное выражение",
			expr:    "2+2*2",
			wantLen: 5,
			wantErr: false,
		},
		{
			name:    "выражение со скобками",
			expr:    "(2+2)*2",
			wantLen: 7,
			wantErr: false,
		},
		{
			name:    "пустое выражение",
			expr:    "",
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "некорректное выражение",
			expr:    "2+",
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "некорректный символ",
			expr:    "2+2@",
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenize(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(tokens) != tt.wantLen {
				t.Errorf("tokenize() got %d tokens, want %d", len(tokens), tt.wantLen)
			}
		})
	}
}

func TestParseExpression(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "простое выражение",
			expr:    "2+2",
			wantErr: false,
		},
		{
			name:    "сложное выражение",
			expr:    "2+2*2",
			wantErr: false,
		},
		{
			name:    "выражение со скобками",
			expr:    "(2+2)*2",
			wantErr: false,
		},
		{
			name:    "пустое выражение",
			expr:    "",
			wantErr: true,
		},
		{
			name:    "некорректное выражение",
			expr:    "2+",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseExpression(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseExpression() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCalc(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		want    float64
		wantErr bool
	}{
		{
			name:    "сложение",
			expr:    "2+2",
			want:    4,
			wantErr: false,
		},
		{
			name:    "вычитание",
			expr:    "5-3",
			want:    2,
			wantErr: false,
		},
		{
			name:    "умножение",
			expr:    "2*3",
			want:    6,
			wantErr: false,
		},
		{
			name:    "деление",
			expr:    "6/2",
			want:    3,
			wantErr: false,
		},
		{
			name:    "приоритет операций",
			expr:    "2+2*2",
			want:    6,
			wantErr: false,
		},
		{
			name:    "скобки",
			expr:    "(2+2)*2",
			want:    8,
			wantErr: false,
		},
		{
			name:    "сложное выражение",
			expr:    "2+2*2+10/2",
			want:    11,
			wantErr: false,
		},
		{
			name:    "деление на ноль",
			expr:    "2/0",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calc(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Calc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Calc() = %v, want %v", got, tt.want)
			}
		})
	}
}
