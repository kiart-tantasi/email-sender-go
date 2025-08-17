package fake

import "time"

func FakeGenerateHtmlTime() {
	// avg time when communicating through kube dns
	time.Sleep(15 * time.Millisecond)
}
