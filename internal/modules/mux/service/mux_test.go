package service
//
// import (
// 	"io"
// 	"log/slog"
// 	"os"
// 	"sync"
// 	"testing"
// )
//
// func TestMux(t *testing.T) {
// 	ctx := t.Context()
//
// 	wg := sync.WaitGroup{}
// 	service := New(slog.New(slog.NewJSONHandler(t.Output(), &slog.HandlerOptions{Level: slog.LevelDebug})), &wg)
//
// 	audioFile, err := os.Open("./unnv-bez-dati.m4a")
// 	if err != nil {
// 		t.Fatalf("opening file: %v", err)
// 	}
// 	defer audioFile.Close()
//
// 	muxState, err := service.Prepare(ctx, audioFile)
// 	if err != nil {
// 		t.Fatalf("preparing mux: %v", err)
// 	}
// 	defer muxState.Close() //nolint:errcheck
//
// 	storage := &mockAudioStorage{
// 		lenghts: map[string]int64{},
// 		wg: &wg,
// 	}
// 	if err = service.Mux(ctx, muxState, storage); err != nil {
// 		t.Fatalf("muxing: %v", err)
// 	}
//
// 	wg.Wait()
//
// 	for filename, written := range storage.lenghts {
// 		t.Logf("%v: %v", filename, written)
// 	}
// }
//
// type mockAudioStorage struct {
// 	lenghts map[string]int64
// 	mu      sync.Mutex
// 	wg      *sync.WaitGroup
// }
//
// func (m *mockAudioStorage) Save(filename string, body io.Reader) {
// 	m.wg.Go(func() {
// 		// buffer := bytes.NewBuffer(nil)
// 		// written, err := io.Copy(buffer, body)
// 		written, err := io.Copy(io.Discard, body)
// 		if err != nil {
// 			slog.Error("COPYING BODY TO DISCARD", "filename", filename, "error", err)
// 			return
// 		}
// 		m.mu.Lock()
// 		m.lenghts[filename] = written
// 		m.mu.Unlock()
// 		// if filename == "playlist.m3u8" {
// 		// 	fmt.Printf("PLAYLIST %s PLAYLIST\n", buffer.Bytes())
// 		// }
// 	})
// }
