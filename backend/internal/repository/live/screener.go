package live

// screener.go defines the candidate universe for each recommendation category
// and the logic for scoring and selecting the top picks using cached market data.
//
// Selection algorithm:
//  1. For each category, score every candidate by 1-month price momentum
//     (% change from 1-month-ago grouped bars to today's grouped bars).
//  2. Sort descending by score; return the top 3.
//  3. Conviction badge is derived from rank: 1st → Strong Buy, 2nd → Buy, 3rd → Speculative
//     (overridden by the curated override field when set).
//
// No additional API calls are made — all price data comes from the grouped
// daily bars already cached in Store.gbCache.

import (
	"fmt"
	"sort"
)

// candidate describes a ticker in the screener universe.
type candidate struct {
	ticker  string
	name    string
	thesis  string
	badge   string // override badge; empty = derive from rank
	recType string // override type; empty = derive from rank
}

// ranked is a scored candidate ready for output.
type ranked struct {
	candidate
	price     float64
	changePct float64 // 24h
	momPct    float64 // 1-month
}

// rankByMomentum scores candidates using 1-month return from cached grouped
// bars, sorts descending, and returns the top n.
func (s *Store) rankByMomentum(candidates []candidate, todayBars, oneMonthBars map[string]float64, n int) []ranked {
	var scored []ranked
	for _, c := range candidates {
		today := todayBars[c.ticker]
		prev := oneMonthBars[c.ticker]
		if today == 0 || prev == 0 {
			continue
		}
		mom := (today - prev) / prev * 100
		scored = append(scored, ranked{candidate: c, price: today, momPct: mom})
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].momPct > scored[j].momPct
	})
	if len(scored) > n {
		scored = scored[:n]
	}
	return scored
}

// rankCrypto scores CoinGecko coins by 24h change and returns the top n.
func rankCrypto(coins []cgCoin, candidates []cgCandidate, n int) []ranked {
	coinMap := make(map[string]cgCoin, len(coins))
	for _, c := range coins {
		coinMap[c.ID] = c
	}
	var scored []ranked
	for _, c := range candidates {
		coin, ok := coinMap[c.cgID]
		if !ok {
			continue
		}
		scored = append(scored, ranked{
			candidate: candidate{
				ticker:  c.ticker,
				name:    coin.Name,
				thesis:  c.thesis,
				badge:   c.badge,
				recType: c.recType,
			},
			price:     coin.CurrentPrice,
			changePct: coin.PriceChangePerc24h,
			momPct:    coin.PriceChangePerc24h,
		})
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].momPct > scored[j].momPct
	})
	if len(scored) > n {
		scored = scored[:n]
	}
	return scored
}

// badgeForRank returns badge + recType based on position (0-indexed).
func badgeForRank(i int, override, typeOverride string) (badge, recType string) {
	if override != "" {
		return override, typeOverride
	}
	switch i {
	case 0:
		return "Strong Buy", "strong-buy"
	case 1:
		return "Buy", "buy"
	default:
		return "Speculative", "speculative"
	}
}

// ── Candidate universes ───────────────────────────────────────────────────────

var aiSemisUniverse = []candidate{
	{ticker: "NVDA", name: "Nvidia Corporation",
		thesis: "AI accelerator dominance with 70-75% market share. Datacenter revenue compounding at scale. Edge devices and software monetization as next catalysts."},
	{ticker: "AVGO", name: "Broadcom Inc",
		thesis: "Custom ASIC leader for Google and Meta hyperscalers. AI revenue growing 70%+ YoY. $1.2T TAM for AI data centres by 2030."},
	{ticker: "AMD", name: "Advanced Micro Devices",
		thesis: "Gaining data centre GPU share against Nvidia. MI300X ramp accelerating. CPU leadership in enterprise and cloud continues."},
	{ticker: "MU", name: "Micron Technology",
		thesis: "HBM memory for AI workloads is a structural growth driver. Supply discipline from industry improving pricing power."},
	{ticker: "AMAT", name: "Applied Materials",
		thesis: "Semiconductor equipment duopoly beneficiary. Gate-all-around and advanced packaging demand driving multi-year capex cycle."},
	{ticker: "MRVL", name: "Marvell Technology",
		thesis: "Custom silicon for cloud hyperscalers. Data infrastructure chips seeing accelerating AI-driven demand."},
	{ticker: "ARM", name: "Arm Holdings",
		thesis: "CPU IP royalties grow with every AI edge device. Licensing model provides recurring revenue with high operating leverage."},
	{ticker: "KLAC", name: "KLA Corporation",
		thesis: "Process control equipment critical to AI chip yields. Near-monopoly in wafer inspection with pricing power."},
	{ticker: "LRCX", name: "Lam Research",
		thesis: "Etch and deposition equipment essential for advanced node transitions. Strong recurring services revenue stream."},
	{ticker: "SMCI", name: "Super Micro Computer",
		thesis: "AI server rack integration at scale. Direct liquid cooling expertise positions it for hyperscaler buildouts."},
}

var disruptorsUniverse = []candidate{
	{ticker: "NBIS", name: "Nebius Group",
		thesis: "AI-focused cloud with 560-620% ARR growth. Full-stack GPU cloud turn-key for AI developers. Fastest projected growth of any public company."},
	{ticker: "MELI", name: "MercadoLibre",
		thesis: "Dominant LatAm e-commerce + fintech. Commerce +40% and fintech +51% YoY. Argentina recovery providing additional tailwind."},
	{ticker: "SOUN", name: "SoundHound AI",
		thesis: "Voice AI platform with accelerating revenue. Down significantly from ATH creating contrarian entry. Automotive and restaurant verticals expanding."},
	{ticker: "PLTR", name: "Palantir Technologies",
		thesis: "AI platform for enterprise and government. AIP product driving commercial acceleration. US government spending tailwind."},
	{ticker: "APP", name: "AppLovin Corporation",
		thesis: "Mobile ad platform with AI-driven targeting. AXON 2.0 driving outsized revenue per impression. Margin expansion story."},
	{ticker: "RKLB", name: "Rocket Lab USA",
		thesis: "Launch + spacecraft manufacturing. Neutron rocket opens medium-lift market. Space economy secular growth."},
	{ticker: "IONQ", name: "IonQ Inc",
		thesis: "Trapped-ion quantum computing leader. Enterprise and government contracts expanding. Early mover in a potential generational technology."},
	{ticker: "HOOD", name: "Robinhood Markets",
		thesis: "Retail brokerage gaining crypto and options volume. New products (credit card, retirement) expanding revenue per user."},
	{ticker: "SOFI", name: "SoFi Technologies",
		thesis: "Digital bank with accelerating lending and financial services. Bank charter enabling cheaper deposit funding. Member cross-sell improving unit economics."},
	{ticker: "UPST", name: "Upstart Holdings",
		thesis: "AI-driven consumer credit platform. Rate sensitivity creates cyclical entry opportunity. Credit model outperforms FICO in default prediction."},
}

var defenseUniverse = []candidate{
	{ticker: "LMT", name: "Lockheed Martin",
		thesis: "F-35 and missile defense programmes provide multi-decade revenue visibility. Direct beneficiary of NATO spending uplift and global conflict escalation."},
	{ticker: "NOC", name: "Northrop Grumman",
		thesis: "B-21 Raider stealth bomber and nuclear modernisation programmes. Space and cyber defence growing contribution."},
	{ticker: "RTX", name: "RTX Corporation",
		thesis: "Patriot missile systems in peak demand. Pratt & Whitney engine services provide recurring revenue. Defense + aerospace diversification."},
	{ticker: "GD", name: "General Dynamics",
		thesis: "Gulfstream jets + combat vehicles + submarines. Backlog at record levels. Reliable dividend grower in uncertain markets."},
	{ticker: "AVAV", name: "AeroVironment",
		thesis: "Loitering munitions and small drone leader. Ukraine conflict proving the ROI of autonomous systems at scale. Switchblade demand structurally elevated."},
	{ticker: "LHX", name: "L3Harris Technologies",
		thesis: "Communications and electronic warfare systems critical to modern conflict. Mission-critical comms upgrades across NATO allies."},
	{ticker: "LDOS", name: "Leidos Holdings",
		thesis: "IT services and intelligence infrastructure for US defence agencies. Sticky government contracts with multi-year visibility."},
	{ticker: "BAH", name: "Booz Allen Hamilton",
		thesis: "Defence consulting and analytics. Classified AI and cyber contracts growing. High renewal rates and mission-critical work."},
	{ticker: "LMB", name: "Limbach Holdings",
		thesis: "MEP infrastructure for hospitals and universities. Oberweis top 2026 pick. Essential infrastructure resilient to macro."},
	{ticker: "HII", name: "Huntington Ingalls",
		thesis: "Sole builder of US nuclear-powered aircraft carriers. Multi-decade shipbuilding backlog. Near-monopoly on critical national security assets."},
}

// cgCandidate is used for crypto picks (scored via CoinGecko data).
type cgCandidate struct {
	ticker  string
	cgID    string
	thesis  string
	badge   string
	recType string
}

var cryptoUniverse = []cgCandidate{
	{ticker: "BTC", cgID: "bitcoin", recType: "buy",
		thesis: "Extreme fear historically marks accumulation phases. Institutional ETF adoption accelerating. Fixed 21M supply cap provides long-term scarcity premium."},
	{ticker: "ETH", cgID: "ethereum", recType: "speculative",
		thesis: "DeFi and L2 ecosystem leader. Staking yield ~4% APY. Standard Chartered $10-40K long-term target. Developer activity remains dominant."},
	{ticker: "SOL", cgID: "solana", recType: "speculative",
		thesis: "High-throughput L1 with growing DeFi TVL. Meme coin and consumer crypto on-chain activity driving fee revenue growth."},
	{ticker: "BNB", cgID: "binancecoin", recType: "speculative",
		thesis: "Binance exchange utility token. BSC DeFi ecosystem second only to Ethereum. Fee burn mechanism reduces supply over time."},
	{ticker: "XRP", cgID: "ripple", recType: "buy",
		thesis: "Cross-border payment settlement with bank partnerships. SEC litigation resolution removed major overhang. Institutional custody expanding."},
}

// fmtMom formats a momentum percentage for display as a stat value.
func fmtMom(pct float64) string {
	return fmt.Sprintf("%+.1f%%", pct)
}
