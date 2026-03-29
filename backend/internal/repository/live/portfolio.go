package live

import "apex-dashboard/backend/internal/models"

// Principal returns static seed data for the Principal.com 401(k) account.
// Update the values below with your actual account data.
func (s *Store) Principal() (models.PortfolioResponse, error) {
	return models.PortfolioResponse{
		Connected: true,
		Summary: models.PortfolioSummary{
			AccountName:  "401(k) Retirement Plan",
			AccountType:  "401(k)",
			Provider:     "Principal Financial Group",
			CurrentValue: "$147,832.14",
			TotalCost:    "$118,500.00",
			TotalGain:    "+$29,332.14",
			TotalGainPct: "+24.75%",
			TotalGainDir: "up",
			WeeklyChange: models.ChangeMetric{
				Amount: "+$412.50",
				Pct:    "+0.28%",
				Dir:    "up",
			},
			MonthlyChange: models.ChangeMetric{
				Amount: "-$2,140.80",
				Pct:    "-1.43%",
				Dir:    "down",
			},
			YTDChange: models.ChangeMetric{
				Amount: "-$6,820.00",
				Pct:    "-4.41%",
				Dir:    "down",
			},
		},
		PerformanceData: models.ChartData{
			Labels: []string{
				"Apr '24", "May '24", "Jun '24", "Jul '24", "Aug '24", "Sep '24",
				"Oct '24", "Nov '24", "Dec '24", "Jan '25", "Feb '25", "Mar '25",
			},
			Data: []float64{
				124500, 126800, 129200, 132400, 130800, 133600,
				137200, 141800, 154652, 155200, 150100, 147832,
			},
		},
		ContribData: models.ChartData{
			Labels: []string{
				"Apr '24", "May '24", "Jun '24", "Jul '24", "Aug '24", "Sep '24",
				"Oct '24", "Nov '24", "Dec '24", "Jan '25", "Feb '25", "Mar '25",
			},
			Data: []float64{
				1200, 1200, 1200, 1200, 1200, 1200,
				1200, 1200, 1200, 1200, 1200, 1200,
			},
		},
		Holdings: []models.PortfolioHolding{
			{Ticker: "VIIIX", Name: "Vanguard Institutional Index", Value: "$53,219.57", Allocation: "36.0%", Gain: "+$9,821.44", GainDir: "up"},
			{Ticker: "VTSNX", Name: "Vanguard Total Intl Stock Idx", Value: "$22,174.82", Allocation: "15.0%", Gain: "+$2,104.20", GainDir: "up"},
			{Ticker: "VBTIX", Name: "Vanguard Total Bond Mkt Idx", Value: "$18,479.02", Allocation: "12.5%", Gain: "-$320.50", GainDir: "down"},
			{Ticker: "VMCIX", Name: "Vanguard Mid-Cap Index", Value: "$17,739.86", Allocation: "12.0%", Gain: "+$1,580.30", GainDir: "up"},
			{Ticker: "VSCIX", Name: "Vanguard Small-Cap Index", Value: "$14,783.21", Allocation: "10.0%", Gain: "+$980.10", GainDir: "up"},
			{Ticker: "VEMPX", Name: "Vanguard Emerging Markets", Value: "$11,086.41", Allocation: "7.5%", Gain: "+$640.80", GainDir: "up"},
			{Ticker: "STABLE", Name: "Principal Stable Value Fund", Value: "$10,349.25", Allocation: "7.0%", Gain: "+$525.70", GainDir: "up"},
		},
		Contributions: []models.Contribution{
			{Date: "Mar 28, 2025", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Feb 28, 2025", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Jan 31, 2025", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Dec 31, 2024", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Nov 29, 2024", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Oct 31, 2024", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Sep 30, 2024", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Aug 30, 2024", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Jul 31, 2024", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Jun 28, 2024", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "May 31, 2024", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
			{Date: "Apr 30, 2024", Amount: "$1,200.00", Type: "employee", Note: "Employee $800 + Employer $400"},
		},
	}, nil
}

// MorganStanley returns static seed data for the Morgan Stanley investment account.
// Update the values below with your actual account data.
func (s *Store) MorganStanley() (models.PortfolioResponse, error) {
	return models.PortfolioResponse{
		Connected: true,
		Summary: models.PortfolioSummary{
			AccountName:  "Investment Portfolio",
			AccountType:  "Brokerage",
			Provider:     "Morgan Stanley",
			CurrentValue: "$62,481.33",
			TotalCost:    "$48,000.00",
			TotalGain:    "+$14,481.33",
			TotalGainPct: "+30.17%",
			TotalGainDir: "up",
			WeeklyChange: models.ChangeMetric{
				Amount: "-$1,243.80",
				Pct:    "-1.95%",
				Dir:    "down",
			},
			MonthlyChange: models.ChangeMetric{
				Amount: "-$3,812.40",
				Pct:    "-5.75%",
				Dir:    "down",
			},
			YTDChange: models.ChangeMetric{
				Amount: "-$5,018.67",
				Pct:    "-7.44%",
				Dir:    "down",
			},
		},
		PerformanceData: models.ChartData{
			Labels: []string{
				"Apr '24", "May '24", "Jun '24", "Jul '24", "Aug '24", "Sep '24",
				"Oct '24", "Nov '24", "Dec '24", "Jan '25", "Feb '25", "Mar '25",
			},
			Data: []float64{
				52400, 53800, 55200, 56900, 55100, 57400,
				58800, 62100, 67500, 68200, 64100, 62481,
			},
		},
		ContribData: models.ChartData{
			Labels: []string{
				"Apr '24", "May '24", "Jun '24", "Jul '24", "Aug '24", "Sep '24",
				"Oct '24", "Nov '24", "Dec '24", "Jan '25", "Feb '25", "Mar '25",
			},
			Data: []float64{
				0, 0, 0, 2000, 0, 0,
				0, 0, 0, 0, 0, 0,
			},
		},
		Holdings: []models.PortfolioHolding{
			{Ticker: "NVDA", Name: "NVIDIA Corporation", Value: "$14,996.72", Allocation: "24.0%", Gain: "+$6,241.30", GainDir: "up"},
			{Ticker: "MSFT", Name: "Microsoft Corporation", Value: "$12,496.27", Allocation: "20.0%", Gain: "+$2,840.10", GainDir: "up"},
			{Ticker: "VTI", Name: "Vanguard Total Stock Market ETF", Value: "$9,372.20", Allocation: "15.0%", Gain: "+$1,372.20", GainDir: "up"},
			{Ticker: "AMZN", Name: "Amazon.com Inc.", Value: "$7,497.76", Allocation: "12.0%", Gain: "+$1,497.76", GainDir: "up"},
			{Ticker: "PLTR", Name: "Palantir Technologies", Value: "$6,248.13", Allocation: "10.0%", Gain: "+$3,748.13", GainDir: "up"},
			{Ticker: "META", Name: "Meta Platforms Inc.", Value: "$6,248.13", Allocation: "10.0%", Gain: "+$1,048.13", GainDir: "up"},
			{Ticker: "BRK.B", Name: "Berkshire Hathaway B", Value: "$5,622.12", Allocation: "9.0%", Gain: "-$265.29", GainDir: "down"},
		},
		Contributions: []models.Contribution{
			{Date: "Jul 15, 2024", Amount: "$2,000.00", Type: "deposit", Note: "Additional investment"},
			{Date: "Jan 3, 2024", Amount: "$10,000.00", Type: "deposit", Note: "Annual contribution"},
			{Date: "Jan 2, 2023", Amount: "$10,000.00", Type: "deposit", Note: "Annual contribution"},
			{Date: "Jan 3, 2022", Amount: "$10,000.00", Type: "deposit", Note: "Annual contribution"},
			{Date: "Jan 4, 2021", Amount: "$10,000.00", Type: "deposit", Note: "Initial investment"},
			{Date: "Jun 15, 2021", Amount: "$8,000.00", Type: "deposit", Note: "Mid-year addition"},
		},
	}, nil
}
