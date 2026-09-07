package activity

import "testing"

func TestStripYouTubeTrackingParams(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "youtu.be strips si",
			in:   "https://youtu.be/-I1aPyp-_uE?si=x",
			want: "https://youtu.be/-I1aPyp-_uE",
		},
		{
			name: "youtube watch keeps v strips si",
			in:   "https://www.youtube.com/watch?v=abc&si=xyz",
			want: "https://www.youtube.com/watch?v=abc",
		},
		{
			name: "keeps playback params",
			in:   "https://www.youtube.com/watch?v=abc&t=30&list=PLfoo&index=2&si=rm",
			want: "https://www.youtube.com/watch?v=abc&t=30&list=PLfoo&index=2",
		},
		{
			name: "strips utm prefix params",
			in:   "https://youtu.be/abc?utm_source=x&utm_medium=y&utm_custom=z&v=abc",
			want: "https://youtu.be/abc?v=abc",
		},
		{
			name: "strips feature pp igsh fbclid gclid",
			in:   "https://youtu.be/abc?feature=share&pp=ygU&igsh=abc&fbclid=def&gclid=ghi",
			want: "https://youtu.be/abc",
		},
		{
			name: "non youtube unchanged",
			in:   "https://example.com/vid?si=keep&utm_source=x",
			want: "https://example.com/vid?si=keep&utm_source=x",
		},
		{
			name: "invalid url unchanged",
			in:   "not-a-url?si=x",
			want: "not-a-url?si=x",
		},
		{
			name: "preserves fragment",
			in:   "https://youtu.be/abc?si=x&t=10#frag",
			want: "https://youtu.be/abc?t=10#frag",
		},
		{
			name: "m subdomain",
			in:   "https://m.youtube.com/watch?v=abc&si=x",
			want: "https://m.youtube.com/watch?v=abc",
		},
		{
			name: "no query unchanged",
			in:   "https://youtu.be/abc",
			want: "https://youtu.be/abc",
		},
		{
			name: "case insensitive param names",
			in:   "https://youtu.be/abc?SI=x&UTM_SOURCE=y&v=abc",
			want: "https://youtu.be/abc?v=abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripYouTubeTrackingParams(tt.in)
			if got != tt.want {
				t.Errorf("StripYouTubeTrackingParams(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
