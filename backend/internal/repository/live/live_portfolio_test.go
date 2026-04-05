package live

import (
	"testing"
)

func TestFidelity_ReturnsConnected(t *testing.T) {
	s := newTestStore(nil)
	resp, err := s.Fidelity()
	if err != nil {
		t.Fatalf("Fidelity() error: %v", err)
	}
	if !resp.Connected {
		t.Error("expected Connected=true")
	}
}

func TestFidelity_HasSummary(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.Fidelity()
	if resp.Summary.Provider != "Fidelity" {
		t.Errorf("Provider = %q, want Fidelity", resp.Summary.Provider)
	}
	if resp.Summary.CurrentValue == "" {
		t.Error("expected non-empty CurrentValue")
	}
}

func TestFidelity_HasHoldings(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.Fidelity()
	if len(resp.Holdings) == 0 {
		t.Error("expected at least one holding")
	}
}

func TestFidelity_HasContributions(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.Fidelity()
	if len(resp.Contributions) == 0 {
		t.Error("expected at least one contribution")
	}
}

func TestFidelity_HasChartData(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.Fidelity()
	if len(resp.PerformanceData.Labels) == 0 {
		t.Error("expected performance chart data")
	}
	if len(resp.ContribData.Labels) == 0 {
		t.Error("expected contribution chart data")
	}
}

func TestSoFi_ReturnsConnected(t *testing.T) {
	s := newTestStore(nil)
	resp, err := s.SoFi()
	if err != nil {
		t.Fatalf("SoFi() error: %v", err)
	}
	if !resp.Connected {
		t.Error("expected Connected=true")
	}
}

func TestSoFi_HasSummary(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.SoFi()
	if resp.Summary.Provider != "SoFi" {
		t.Errorf("Provider = %q, want SoFi Invest", resp.Summary.Provider)
	}
}

func TestSoFi_HasHoldings(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.SoFi()
	if len(resp.Holdings) == 0 {
		t.Error("expected at least one holding")
	}
}

func TestSoFi_HasChartData(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.SoFi()
	if len(resp.PerformanceData.Labels) == 0 {
		t.Error("expected performance chart data")
	}
}

func TestPrincipal_ReturnsConnected(t *testing.T) {
	s := newTestStore(nil)
	resp, err := s.Principal()
	if err != nil {
		t.Fatalf("Principal() error: %v", err)
	}
	if !resp.Connected {
		t.Error("expected Connected=true")
	}
}

func TestPrincipal_HasSummary(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.Principal()
	if resp.Summary.Provider != "Principal Financial Group" {
		t.Errorf("Provider = %q, want Principal Financial Group", resp.Summary.Provider)
	}
}

func TestPrincipal_HasHoldings(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.Principal()
	if len(resp.Holdings) == 0 {
		t.Error("expected at least one holding")
	}
}

func TestPrincipal_HasContributions(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.Principal()
	if len(resp.Contributions) == 0 {
		t.Error("expected at least one contribution")
	}
}

func TestMorganStanley_ReturnsConnected(t *testing.T) {
	s := newTestStore(nil)
	resp, err := s.MorganStanley()
	if err != nil {
		t.Fatalf("MorganStanley() error: %v", err)
	}
	if !resp.Connected {
		t.Error("expected Connected=true")
	}
}

func TestMorganStanley_HasSummary(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.MorganStanley()
	if resp.Summary.Provider != "Morgan Stanley" {
		t.Errorf("Provider = %q, want Morgan Stanley", resp.Summary.Provider)
	}
}

func TestMorganStanley_HasHoldings(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.MorganStanley()
	if len(resp.Holdings) == 0 {
		t.Error("expected at least one holding")
	}
}

func TestMorganStanley_HasChartData(t *testing.T) {
	s := newTestStore(nil)
	resp, _ := s.MorganStanley()
	if len(resp.PerformanceData.Labels) == 0 {
		t.Error("expected performance chart data")
	}
}
