package pricer

type Pricer interface {
	CalculateTripCost(weightClass, distance int, earnWithFuel bool) int
	CalculateTripRevenue(tripCost int) int
}
