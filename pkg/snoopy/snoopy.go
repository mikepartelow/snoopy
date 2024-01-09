package snoopy

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Snoopy struct {
	logger *slog.Logger
	snoops []snoop
}

func New(filename string, logger *slog.Logger) (*Snoopy, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config %q: %w", filename, err)
	}
	defer file.Close()

	var snoops []snoop
	err = yaml.NewDecoder(file).Decode(&snoops)
	if err != nil {
		return nil, fmt.Errorf("error decoding config %q: %w", filename, err)
	}

	for i := 0; i < len(snoops); i++ {
		snoops[i].logger = logger.With("local", snoops[i].Local, "upstream", snoops[i].Upstream)
	}

	return &Snoopy{
		logger: logger,
		snoops: snoops,
	}, nil
}

func (s *Snoopy) Run() {
	var wg sync.WaitGroup

	for _, snoop := range s.snoops {
		snoop := snoop

		wg.Add(1)
		go func() {
			defer wg.Done()

			snoop.logger.Debug("Starting snoop server", "local", snoop.Local, "upstream", snoop.Upstream, "logfile", snoop.Logfile)
			if err := http.ListenAndServe(snoop.Local, &snoop); err != nil {
				snoop.logger.Error("ListenAndServe:", err)
				panic(err)
			}
		}()
	}

	wg.Wait()
}
