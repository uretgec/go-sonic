package sonic

import (
	"context"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func findDummyIndex(words string) int {
	return strings.Index(sanitize(words), " ") // -1 not found
}

func sanitize(text string) string {
	r := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\"", "\\\"", "'", `\'`)
	text = r.Replace(text)

	return text
}

func truncateText(str string, limitter int) string {
	if utf8.RuneCountInString(str) < limitter {
		return str
	}

	return string([]rune(str)[:limitter])
}

func Sleep(ctx context.Context, dur time.Duration) error {
	t := time.NewTimer(dur)
	defer t.Stop()

	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// A Once will perform a successful action exactly once.
//
// Unlike a sync.Once, this Once's func returns an error
// and is re-armed on failure.
type Once struct {
	m    sync.Mutex
	done uint32
}

// Do calls the function f if and only if Do has not been invoked
// without error for this instance of Once.  In other words, given
// 	var once Once
// if once.Do(f) is called multiple times, only the first call will
// invoke f, even if f has a different value in each invocation unless
// f returns an error.  A new instance of Once is required for each
// function to execute.
//
// Do is intended for initialization that must be run exactly once.  Since f
// is niladic, it may be necessary to use a function literal to capture the
// arguments to a function to be invoked by Do:
// 	err := config.once.Do(func() error { return config.init(filename) })
func (o *Once) Do(f func() error) error {
	if atomic.LoadUint32(&o.done) == 1 {
		return nil
	}
	// Slow-path.
	o.m.Lock()
	defer o.m.Unlock()
	var err error
	if o.done == 0 {
		err = f()
		if err == nil {
			atomic.StoreUint32(&o.done, 1)
		}
	}
	return err
}

func RetryBackoff(retry int, minBackoff, maxBackoff time.Duration) time.Duration {
	if retry < 0 {
		panic("not reached")
	}
	if minBackoff == 0 {
		return 0
	}

	d := minBackoff << uint(retry)
	if d < minBackoff {
		return maxBackoff
	}

	d = minBackoff + time.Duration(Int63n(int64(d)))

	if d > maxBackoff || d < minBackoff {
		d = maxBackoff
	}

	return d
}

// Rand
// Int returns a non-negative pseudo-random int.
func Int() int { return pseudo.Int() }

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func Intn(n int) int { return pseudo.Intn(n) }

// Int63n returns, as an int64, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func Int63n(n int64) int64 { return pseudo.Int63n(n) }

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers [0,n).
func Perm(n int) []int { return pseudo.Perm(n) }

// Seed uses the provided seed value to initialize the default Source to a
// deterministic state. If Seed is not called, the generator behaves as if
// seeded by Seed(1).
func Seed(n int64) { pseudo.Seed(n) }

var pseudo = rand.New(&source{src: rand.NewSource(1)})

type source struct {
	src rand.Source
	mu  sync.Mutex
}

func (s *source) Int63() int64 {
	s.mu.Lock()
	n := s.src.Int63()
	s.mu.Unlock()
	return n
}

func (s *source) Seed(seed int64) {
	s.mu.Lock()
	s.src.Seed(seed)
	s.mu.Unlock()
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements.
// swap swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) { pseudo.Shuffle(n, swap) }
