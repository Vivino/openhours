package openhours

import (
	"reflect"
	"testing"
	"time"
)

var l *time.Location

func init() {
	var err error
	l, err = time.LoadLocation("Europe/London") // is a goot test example since i know when the two clock change occur
	if err != nil {
		panic("could not load location")
	}
}

func Test_cleanStr(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{"capital letters", "Mo-Fr 10:00-12:00,12:30-16:00", "mo-fr 10:00-12:00,12:30-16:00"},
		{"comma after space", "Mo-Fr 10:00-12:00, 12:30-16:00", "mo-fr 10:00-12:00,12:30-16:00"},
		{"comma before space", "Mo-Fr 10:00-12:00 ,12:30-16:00", "mo-fr 10:00-12:00,12:30-16:00"},
		{"comma before and after space", "Mo-Fr 10:00-12:00 , 12:30-16:00", "mo-fr 10:00-12:00,12:30-16:00"},
		{"front space", " Mo-Fr 10:00-12:00,12:30-16:00", "mo-fr 10:00-12:00,12:30-16:00"},
		{"trailing space", "Mo-Fr 10:00-12:00,12:30-16:00 ", "mo-fr 10:00-12:00,12:30-16:00"},
		{"both spaces", " Mo-Fr 10:00-12:00,12:30-16:00 ", "mo-fr 10:00-12:00,12:30-16:00"},
		{"mixed both spaces/tabs", " 	Mo-Fr 10:00-12:00,12:30-16:00	 ", "mo-fr 10:00-12:00,12:30-16:00"},
		{"inner mixed spaces/tabs", " 	 	Mo-Fr 	 10:00-12:00,12:30-16:00 	 	 ", "mo-fr 10:00-12:00,12:30-16:00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanStr(tt.args); got != tt.want {
				t.Errorf("cleanStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_simplifyDays(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []int
	}{
		{"simple", "mo", []int{1}},
		{"double with error", "mo,mardi", []int{1}},
		{"double with error", "mo,mardi", []int{1}},
		{"double", "we,fr", []int{3, 5}},
		{"range", "we-fr", []int{3, 4, 5}},
		{"range with double", "mo,we-fr,su", []int{0, 1, 3, 4, 5}},
		{"error -", "mo-pl", []int{}},
		{"error ,", "pl,mo", []int{1}},
		{"weird range", "fr-mo", []int{0, 1, 5, 6}},
		{"dupicate days", "mo-tu,tu,tu-fr,fr", []int{1, 2, 3, 4, 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := simplifyDays(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("simplifyDays() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_simplifyHour(t *testing.T) {
	tests := []struct {
		args  string
		want  int
		want1 int
		want2 int
	}{
		{"00:00", 0, 0, 0},
		{"00:00:00", 0, 0, 0},
		{"00:00:05", 0, 0, 5},
		{"10:30", 10, 30, 0},
		{"09:05", 9, 5, 0},
		{"24:00", 24, 0, 0},
		{"00:-10", 0, 0, 0},
		{"24:01", 0, 0, 0},
		{"-50:99", 0, 0, 0},
		{"33:33:33", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			got, got1, got2 := simplifyTime(tt.args)
			if got != tt.want {
				t.Errorf("simplifyHour() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("simplifyHour() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("simplifyHour() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_feature_simple(t *testing.T) {
	o := New("mo 08:00-18:00", l)
	tests := []struct {
		args time.Time
		want bool
	}{
		{time.Date(2019, 3, 4, 8, 0, 0, 0, l), true}, // special case start = true
		{time.Date(2019, 3, 4, 17, 59, 0, 0, l), true},
		{time.Date(2019, 3, 4, 18, 0, 0, 0, l), false}, // special case end = false
		{time.Date(2019, 3, 4, 7, 0, 0, 0, l), false},
		{time.Date(2019, 3, 4, 19, 0, 0, 0, l), false},
	}
	for _, tt := range tests {
		t.Run(tt.args.String(), func(t *testing.T) {
			if got := o.Match(tt.args); got != tt.want {
				t.Errorf("simplifyHour() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_feature_two(t *testing.T) {
	o := New("mo 08:00-12:00,13:00-17:00", l)
	tests := []struct {
		args time.Time
		want bool
	}{
		{time.Date(2019, 3, 4, 8, 0, 0, 0, l), true}, // special case start = true
		{time.Date(2019, 3, 4, 9, 0, 0, 0, l), true},
		{time.Date(2019, 3, 4, 13, 0, 0, 0, l), true}, // special case start = true
		{time.Date(2019, 3, 4, 15, 0, 0, 0, l), true},
		{time.Date(2019, 3, 4, 12, 30, 0, 0, l), false}, // between
		{time.Date(2019, 3, 4, 17, 59, 0, 0, l), false},
		{time.Date(2019, 3, 4, 17, 0, 0, 0, l), false}, // special case end = false
		{time.Date(2019, 3, 4, 12, 0, 0, 0, l), false}, // special case end = false
		{time.Date(2019, 3, 4, 7, 0, 0, 0, l), false},
		{time.Date(2019, 3, 4, 19, 0, 0, 0, l), false},
	}
	for _, tt := range tests {
		t.Run(tt.args.String(), func(t *testing.T) {
			if got := o.Match(tt.args); got != tt.want {
				t.Errorf("simplifyHour() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenHours_NextDur(t *testing.T) {
	o := New("mo 08:00-18:00", l)
	tests := []struct {
		name  string
		args  time.Time
		want  bool
		want1 time.Duration
	}{
		{"1 hour before start", time.Date(2019, 3, 4, 7, 0, 0, 0, l), false, time.Hour},
		{"at start", time.Date(2019, 3, 4, 8, 0, 0, 0, l), true, 10 * time.Hour},
		{"1 hour after start", time.Date(2019, 3, 4, 9, 0, 0, 0, l), true, 9 * time.Hour},
		{"1 hour before end", time.Date(2019, 3, 4, 17, 0, 0, 0, l), true, time.Hour},
		{"at end", time.Date(2019, 3, 4, 18, 0, 0, 0, l), false, time.Hour*24*7 - time.Hour*10},
		{"1 day after start (closed)", time.Date(2019, 3, 5, 8, 0, 0, 0, l), false, time.Hour * 24 * 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := o.NextDur(tt.args)
			if got != tt.want {
				t.Errorf("OpenHours.NextDur() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("OpenHours.NextDur() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestOpenHours_Special_NextDur(t *testing.T) {
	o := New("su 03:00-05:00", l)
	tests := []struct {
		name  string
		args  time.Time
		want  bool
		want1 time.Duration
	}{
		{"2 h before (3 if there was no clock change)", time.Date(2019, 3, 31, 0, 0, 0, 0, l), false, time.Hour * 2},
		{"4 h before (3 if there was no clock change)", time.Date(2019, 10, 27, 0, 0, 0, 0, l), false, time.Hour * 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := o.NextDur(tt.args)
			if got != tt.want {
				t.Errorf("OpenHours.NextDur() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("OpenHours.NextDur() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		args  string
		args2 *time.Location
		want  OpenHours
	}{
		{"empty", "", l, []time.Time{newDate(0, 0, 0, 0, 0, l), newDate(7, 0, 0, 0, 0, l)}},
		{"empty ;", ";", l, []time.Time{newDate(0, 0, 0, 0, 0, l), newDate(7, 0, 0, 0, 0, l)}},
		{"all day ;", "su-sa 00:00-24:00;", l, []time.Time{newDate(0, 0, 0, 0, 0, l), newDate(7, 0, 0, 0, 0, l)}},
		{"empty and no tz", "", nil, []time.Time{newDate(0, 0, 0, 0, 0, time.UTC), newDate(7, 0, 0, 0, 0, time.UTC)}},
		{"order on same sentense", "mo,tu 10:00-11:00", nil, New("tu,mo 10:00-11:00", nil)},
		{"order on different sentenses", "mo 10:00-11:00;tu 10:00-12:00", nil, New("tu 10:00-12:00;mo 10:00-11:00", nil)},
		{"complex = simple", "su-sa 00:00-12:00,12:00-24:00", l, New("", l)},
		{"complex = simple", "su-sa 00:00-12:00;su-sa 12:00-24:00", l, New("", l)},
		{"time windows order does not matter anymore", "mo-su 00:00-24:00", l, New("", l)},
		{"weird times in same sentence", "mo-fr 00:00-15:00,10:00-24:00", l, New("mo-fr 00:00-24:00", l)},
		{"weird times in same sentence one contained", "mo-fr 10:00-15:00,00:00-24:00", l, New("mo-fr 00:00-24:00", l)},
		{"weird times in different sentence", "mo-fr 00:00-15:00;mo-fr 10:00-24:00", l, New("mo-fr 00:00-24:00", l)},
		{"weird times in different sentence contained", "mo-fr 10:00-15:00;mo-fr 00:00-24:00", l, New("mo-fr 00:00-24:00", l)},
		{"total chaos", "tu-fr 10:00-15:00;mo 08:00-09:00;mo-fr 00:00-24:00", l, New("mo-fr 00:00-24:00", l)},
		{"one day", "mo 10:00-15:00", l, []time.Time{newDate(1, 10, 0, 0, 0, l), newDate(1, 15, 0, 0, 0, l)}},
		{"two days", "mo 10:00-15:00;fr 08:00-14:00", l, []time.Time{newDate(1, 10, 0, 0, 0, l), newDate(1, 15, 0, 0, 0, l), newDate(5, 8, 0, 0, 0, l), newDate(5, 14, 0, 0, 0, l)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args, tt.args2)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenHours_NextDate(t *testing.T) {
	o := New("su 03:00-05:00", l)
	tests := []struct {
		name  string
		args  time.Time
		want  bool
		want1 time.Time
	}{
		{"2 h before (3 if there was no clock change)", time.Date(2019, 3, 31, 0, 0, 0, 0, l), false, time.Date(2019, 3, 31, 3, 0, 0, 0, l)},
		{"4 h before (3 if there was no clock change)", time.Date(2019, 10, 27, 0, 0, 0, 0, l), false, time.Date(2019, 10, 27, 3, 0, 0, 0, l)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := o.NextDate(tt.args)
			if got != tt.want {
				t.Errorf("OpenHours.NextDate() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("OpenHours.NextDate() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func pDate(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) *time.Time {
	t := time.Date(year, month, day, hour, min, sec, nsec, loc)
	return &t
}

func TestOpenHours_When(t *testing.T) {
	type args struct {
		t time.Time
		d time.Duration
	}
	tests := []struct {
		name string
		o    OpenHours
		args args
		want *time.Time
	}{
		{"at start of open and have time", New("mo 10:00-15:00", l), args{time.Date(2019, 3, 11, 10, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 11, 10, 0, 0, 0, l)},
		{"before start of open and have time", New("mo 10:00-15:00", l), args{time.Date(2019, 3, 11, 9, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 11, 10, 0, 0, 0, l)},
		{"at end of open and have time", New("mo 10:00-15:00", l), args{time.Date(2019, 3, 11, 15, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 18, 10, 0, 0, 0, l)},
		{"after end of open and have time", New("mo 10:00-15:00", l), args{time.Date(2019, 3, 11, 16, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 18, 10, 0, 0, 0, l)},
		{"between open and have time", New("mo 10:00-15:00", l), args{time.Date(2019, 3, 11, 11, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 11, 11, 0, 0, 0, l)},
		{"between open and no time", New("mo 10:00-15:00", l), args{time.Date(2019, 3, 11, 14, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 18, 10, 0, 0, 0, l)},
		{"no time", New("mo 10:00-11:00", l), args{time.Date(2019, 3, 11, 14, 0, 0, 0, l), time.Hour * 4}, nil},
		{"at start of open and have time +fri", New("mo 10:00-15:00;fr 08:00-14:00", l), args{time.Date(2019, 3, 11, 10, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 11, 10, 0, 0, 0, l)},
		{"before start of open and have time +fri", New("mo 10:00-15:00;fr 08:00-14:00", l), args{time.Date(2019, 3, 11, 9, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 11, 10, 0, 0, 0, l)},
		{"at end of open and have time +fri", New("mo 10:00-15:00;fr 08:00-14:00", l), args{time.Date(2019, 3, 11, 15, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 15, 8, 0, 0, 0, l)},
		{"after end of open and have time +fri", New("mo 10:00-15:00;fr 08:00-14:00", l), args{time.Date(2019, 3, 11, 16, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 15, 8, 0, 0, 0, l)},
		{"between open and have time +fri", New("mo 10:00-15:00;fr 08:00-14:00", l), args{time.Date(2019, 3, 11, 11, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 11, 11, 0, 0, 0, l)},
		{"between open and no time +fri", New("mo 10:00-15:00;fr 08:00-14:00", l), args{time.Date(2019, 3, 11, 14, 0, 0, 0, l), time.Hour * 4}, pDate(2019, 3, 15, 8, 0, 0, 0, l)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.When(tt.args.t, tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OpenHours.When() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenHours_Add(t *testing.T) {
	type args struct {
		t time.Time
		d time.Duration
	}
	tests := []struct {
		name string
		o    OpenHours
		args args
		want *time.Time
	}{
		{
			"at start of open and have time",
			OpenHours{}.Add(time.Date(2019, 3, 11, 10, 0, 0, 0, l), time.Date(2019, 3, 11, 10, 30, 0, 0, l)), // mo 10:00-10:30
			args{time.Date(2019, 3, 11, 9, 0, 0, 0, l), time.Second}, pDate(2019, 3, 11, 10, 0, 0, 0, l),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.When(tt.args.t, tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OpenHours.When() = %v, want %v", got, tt.want)
			}
		})
	}
}
