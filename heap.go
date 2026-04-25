package pprof

import (
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
)

func registerHeapHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/debug/heap/gc", heapGC)
	mux.HandleFunc("/debug/heap/free", heapFree)
	mux.HandleFunc("/debug/heap/gcp", heapGCPercent)
	mux.HandleFunc("/debug/heap/memlimit", heapMemLimit)
}

func heapGC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	runtime.GC()
	fmt.Fprintln(w, "GC triggered")
}

func heapFree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	debug.FreeOSMemory()
	fmt.Fprintln(w, "FreeOSMemory triggered")
}

func heapGCPercent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	v, err := strconv.Atoi(r.URL.Query().Get("v"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid v: %v\n", err)
		return
	}
	old := debug.SetGCPercent(v)
	fmt.Fprintf(w, "SetGCPercent %d -> %d\n", old, v)
}

func heapMemLimit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	v, err := strconv.ParseInt(r.URL.Query().Get("v"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid v: %v\n", err)
		return
	}
	old := debug.SetMemoryLimit(v)
	fmt.Fprintf(w, "SetMemoryLimit %d -> %d\n", old, v)
}
