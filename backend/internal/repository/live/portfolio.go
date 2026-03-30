package live

import "apex-dashboard/backend/internal/models"

// Fidelity returns static seed data for the Fidelity brokerage account.
// Update the values below with your actual account data.
func (s *Store) Fidelity() (models.PortfolioResponse, error) {
	return models.PortfolioResponse{
		Connected: true,
		Summary: models.PortfolioSummary{
			AccountName:  "Individual Brokerage",
			AccountType:  "Brokerage",
			Provider:     "Fidelity",
			CurrentValue: "$43,218.74",
			TotalCost:    "$35,000.00",
			TotalGain:    "+$8,218.74",
			TotalGainPct: "+23.48%",
			TotalGainDir: "up",
			WeeklyChange: models.ChangeMetric{
				Amount: "-$612.30",
				Pct:    "-1.40%",
				Dir:    "down",
			},
			MonthlyChange: models.ChangeMetric{
				Amount: "-$1,840.50",
				Pct:    "-4.08%",
				Dir:    "down",
			},
			YTDChange: models.ChangeMetric{
				Amount: "-$3,210.00",
				Pct:    "-6.91%",
				Dir:    "down",
			},
		},
		PerformanceData: models.ChartData{
			Labels: []string{
				"Apr '24", "May '24", "Jun '24", "Jul '24", "Aug '24", "Sep '24",
				"Oct '24", "Nov '24", "Dec '24", "Jan '25", "Feb '25", "Mar '25",
			},
			Data: []float64{
				37200, 38100, 39400, 40800, 39600, 41200,
				42100, 44600, 46428, 46100, 44900, 43218,
			},
		},
		ContribData: models.ChartData{
			Labels: []string{
				"Apr '24", "May '24", "Jun '24", "Jul '24", "Aug '24", "Sep '24",
				"Oct '24", "Nov '24", "Dec '24", "Jan '25", "Feb '25", "Mar '25",
			},
			Data: []float64{
				0, 0, 0, 0, 0, 0,
				0, 0, 0, 5000, 0, 0,
			},
		},
		Holdings: []models.PortfolioHolding{
			{Ticker: "FSKAX", Name: "Fidelity Total Market Index", Value: "$13,829.99", Allocation: "32.0%", Gain: "+$2,829.99", GainDir: "up"},
			{Ticker: "FXAIX", Name: "Fidelity 500 Index Fund", Value: "$10,804.69", Allocation: "25.0%", Gain: "+$2,304.69", GainDir: "up"},
			{Ticker: "FTIHX", Name: "Fidelity Total Intl Index", Value: "$6,482.81", Allocation: "15.0%", Gain: "+$982.81", GainDir: "up"},
			{Ticker: "NVDA", Name: "NVIDIA Corporation", Value: "$5,186.25", Allocation: "12.0%", Gain: "+$2,686.25", GainDir: "up"},
			{Ticker: "AAPL", Name: "Apple Inc.", Value: "$3,457.50", Allocation: "8.0%", Gain: "+$457.50", GainDir: "up"},
			{Ticker: "FZROX", Name: "Fidelity ZERO Total Market", Value: "$2,160.94", Allocation: "5.0%", Gain: "+$160.94", GainDir: "up"},
			{Ticker: "FBTC", Name: "Fidelity Wise Origin Bitcoin", Value: "$1,296.56", Allocation: "3.0%", Gain: "-$203.44", GainDir: "down"},
		},
		Contributions: []models.Contribution{
			{Date: "Jan 2, 2025", Amount: "$5,000.00", Type: "deposit", Note: "Annual contribution"},
			{Date: "Jan 3, 2024", Amount: "$5,000.00", Type: "deposit", Note: "Annual contribution"},
			{Date: "Jan 4, 2023", Amount: "$5,000.00", Type: "deposit", Note: "Annual contribution"},
			{Date: "Jan 5, 2022", Amount: "$5,000.00", Type: "deposit", Note: "Annual contribution"},
			{Date: "Jun 15, 2022", Amount: "$5,000.00", Type: "deposit", Note: "Mid-year addition"},
			{Date: "Jun 10, 2023", Amount: "$5,000.00", Type: "deposit", Note: "Mid-year addition"},
			{Date: "Jul 1, 2024", Amount: "$5,000.00", Type: "deposit", Note: "Mid-year addition"},
		},
	}, nil
}

// SoFi returns static seed data for the SoFi Invest brokerage account.
// Update the values below with your actual account data.
func (s *Store) SoFi() (models.PortfolioResponse, error) {
	return models.PortfolioResponse{
		Connected: true,
		Summary: models.PortfolioSummary{
			AccountName:  "Active Invest",
			AccountType:  "Brokerage",
			Provider:     "SoFi",
			CurrentValue: "$8,142.60",
			TotalCost:    "$7,500.00",
			TotalGain:    "+$642.60",
			TotalGainPct: "+8.57%",
			TotalGainDir: "up",
			WeeklyChange: models.ChangeMetric{
				Amount: "-$184.20",
				Pct:    "-2.21%",
				Dir:    "down",
			},
			MonthlyChange: models.ChangeMetric{
				Amount: "-$390.80",
				Pct:    "-4.58%",
				Dir:    "down",
			},
			YTDChange: models.ChangeMetric{
				Amount: "-$607.40",
				Pct:    "-6.94%",
				Dir:    "down",
			},
		},
		PerformanceData: models.ChartData{
			Labels: []string{
				"Apr '24", "May '24", "Jun '24", "Jul '24", "Aug '24", "Sep '24",
				"Oct '24", "Nov '24", "Dec '24", "Jan '25", "Feb '25", "Mar '25",
			},
			Data: []float64{
				6800, 7050, 7280, 7540, 7310, 7620,
				7880, 8410, 8750, 8720, 8430, 8142,
			},
		},
		ContribData: models.ChartData{
			Labels: []string{
				"Apr '24", "May '24", "Jun '24", "Jul '24", "Aug '24", "Sep '24",
				"Oct '24", "Nov '24", "Dec '24", "Jan '25", "Feb '25", "Mar '25",
			},
			Data: []float64{
				250, 250, 250, 250, 250, 250,
				250, 250, 250, 250, 250, 250,
			},
		},
		Holdings: []models.PortfolioHolding{
			{Ticker: "VTI", Name: "Vanguard Total Stock Market ETF", Value: "$2,442.78", Allocation: "30.0%", Gain: "+$192.78", GainDir: "up"},
			{Ticker: "QQQ", Name: "Invesco QQQ Trust", Value: "$1,628.52", Allocation: "20.0%", Gain: "+$128.52", GainDir: "up"},
			{Ticker: "PLTR", Name: "Palantir Technologies", Value: "$1,221.39", Allocation: "15.0%", Gain: "+$471.39", GainDir: "up"},
			{Ticker: "ARKK", Name: "ARK Innovation ETF", Value: "$814.26", Allocation: "10.0%", Gain: "-$185.74", GainDir: "down"},
			{Ticker: "SOFI", Name: "SoFi Technologies Inc.", Value: "$814.26", Allocation: "10.0%", Gain: "-$85.74", GainDir: "down"},
			{Ticker: "IONQ", Name: "IonQ Inc.", Value: "$610.70", Allocation: "7.5%", Gain: "+$110.70", GainDir: "up"},
			{Ticker: "RKLB", Name: "Rocket Lab USA", Value: "$610.69", Allocation: "7.5%", Gain: "+$10.69", GainDir: "up"},
		},
		Contributions: []models.Contribution{
			{Date: "Mar 1, 2025", Amount: "$250.00", Type: "deposit", Note: "Monthly auto-invest"},
			{Date: "Feb 1, 2025", Amount: "$250.00", Type: "deposit", Note: "Monthly auto-invest"},
			{Date: "Jan 1, 2025", Amount: "$250.00", Type: "deposit", Note: "Monthly auto-invest"},
			{Date: "Dec 1, 2024", Amount: "$250.00", Type: "deposit", Note: "Monthly auto-invest"},
			{Date: "Nov 1, 2024", Amount: "$250.00", Type: "deposit", Note: "Monthly auto-invest"},
			{Date: "Oct 1, 2024", Amount: "$250.00", Type: "deposit", Note: "Monthly auto-invest"},
		},
	}, nil
}

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
