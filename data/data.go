package data

type RoomType struct {
	RoomName string
	BedType  string
	RoomSize string
	Tags     []string
}

type RoomAttribute struct {
	IncludeBreakfast bool
	FreeCancelation  bool
}

type Availability struct {
	RoomType   RoomType
	Price      float64
	Attributes RoomAttribute
}

type CategoryReview struct {
	Name  string
	Score float64
}

type Review struct {
	AllScore        float64
	AllCount        int64
	CategoryReviews []CategoryReview
}

type AccomodationDetail struct {
	Name           string
	Address        string
	Tags           []string
	Description    string
	Availabilities []Availability
}
