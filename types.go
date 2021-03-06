package main

import (
	"image"
	"log"
	"time"
)

/*
 * Facebook Graph API structs for retrieving page photos
 */

//https://developers.facebook.com/docs/graph-api/reference/page/photos/
//?type=uploaded
type PhotosEdge struct {
	Data  []PhotosEdgePhoto  `json:"data"`
	Error GraphErrorResponse `json:"error"`
}

type PhotosEdgePhoto struct {
	CreatedTime string `json:"created_time"`
	UpdatedTime string `json:"updated_time"`
	Name        string `json:"name"`
	Id          string `json:"id"`
}

//https://developers.facebook.com/docs/graph-api/reference/photo/
//?fields=images
type PhotoNode struct {
	Images []PhotoNodeImage   `json:"images"`
	Error  GraphErrorResponse `json:"error"`
}

type PhotoNodeImage struct {
	Source string `json:"source"`
}

//Facebook Graph API page album edge
//https://developers.facebook.com/docs/graph-api/reference/page/albums
type AlbumsEdge struct {
	Data  []AlbumsEdgeAlbum  `json:"data"`
	Error GraphErrorResponse `json:"error"`
}

type AlbumsEdgeAlbum struct {
	CreatedTime string `json:"created_time"`
	UpdatedTime string `json:"updated_time"`
	Name        string `json:"name"`
	Id          string `json:"id"`
}

//Facebook Graph API page info
//https://developers.facebook.com/docs/graph-api/reference/page/
type PageInfoEdge struct {
	Phone       string   `json:"phone"`
	Emails      []string `json:"emails"`
	GeneralInfo string   `json:"general_info"`

	Error GraphErrorResponse `json:"error"`
}

//Facebook Graph API standard error response
//https://developers.facebook.com/docs/graph-api/using-graph-api#errors
type GraphErrorResponse struct {
	Message          string `json:"message"`
	Type             string `json:"type"`
	Code             int    `json:"code"`
	Error_Subcode    int    `json:"error_subcode"`
	Error_User_Title string `json:"error_user_title"`
	Error_User_Msg   string `json:"error_user_msg"`
	FBTrace_Id       string `json:"fbtrace_id"`
}

/*
type Departure struct {
	RollCall    time.Time `json:"rollCall"`
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
	SeatCount   int       `json:"seatCount"`
	SeatType    string    `json:"seatType"`
	Canceled    bool      `json:"canceled"`
	PhotoSource string    `json:"photoSource"`
}
*/

//Returned when searching all terminal keywords in plaintext
type TerminalKeywordsResult struct {
	Keyword  string
	Distance int
}

//Represents location map in terminal/location keyword file
type TerminalLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

//Terminal representation
//Used for both Terminal list and keywords list depending on which files loaded from.
type Terminal struct {
	Title    string           `json:"title"`
	Id       string           `json:"id"`
	Keywords []string         `json:"keywords"`
	Location TerminalLocation `json:"location"`
	Timezone *time.Location

	PageInfoEdge

	TimezoneOffset int    `json:"tzOffset"` //Used to debug TZ offset export
	TimezoneTitle string `json:"tzTitle"`
}

//Processed version of downloaded photo
type Slide struct {
	SaveType      SaveImageType
	Suffix        string
	Extension     string
	Terminal      Terminal
	FBNodeId      string
	FBCreatedTime time.Time

	PlainText string
	HOCRText  string
}

/*
 * Representation of data in an image
 */

type SharedInfo struct {
	BBox image.Rectangle
}

//Destination representation
type Destination struct {
	SharedInfo
	TerminalTitle    string
	Spelling         string
	SpellingDistance int

	//RollCall for the Destination
	//non nil value indicates 'anchor' (same horizonal level RollCall) Destination for grouping Destinations to other nearby Destinations.
	//later linked for all Destinations in Grouping for insertion of flight to database
	LinkedRollCall *RollCall
	//SeatsAvailable for Destination.
	//NOT used to indicate 'anchor'. Use LinkedRollCall to indicate anchor.
	LinkedSeatsAvailable *SeatsAvailable
}

//RollCall representation
type RollCall struct {
	Time time.Time
	SharedInfo

	//SeatsAvailabe that is in the same row as RollCall
	LinkedSeatsAvailable *SeatsAvailable
}

//SeatsAvailable representation
type SeatsAvailable struct {
	Number int
	Letter string
	SharedInfo
}

//"Grouping" of multiples Destinations for single RollCall/SeatsAvailable
type Grouping struct {
	Destinations []Destination

	//non nil value indicates Grouping contains 'anchor' Destination
	LinkedRollCall *RollCall
	SharedInfo
}

//Update Grouping struct BBox to include all Destinations in grouping
func (g *Grouping) updateBBox() {
	if len((*g).Destinations) == 0 {
		log.Fatal("Grouping empty. Cannot updateBBox")
	}

	(*g).BBox = (*g).Destinations[0].BBox

	for _, d := range (*g).Destinations {
		if d.BBox.Min.X < (*g).BBox.Min.X {
			(*g).BBox.Min.X = d.BBox.Min.X
		}
		if d.BBox.Min.Y < (*g).BBox.Min.Y {
			(*g).BBox.Min.Y = d.BBox.Min.Y
		}
		if d.BBox.Max.X > (*g).BBox.Max.X {
			(*g).BBox.Max.X = d.BBox.Max.X
		}
		if d.BBox.Max.Y > (*g).BBox.Max.Y {
			(*g).BBox.Max.Y = d.BBox.Max.Y
		}
	}
}

//Representation of a specific flight
type Flight struct {
	Origin              string    `json:"origin"`
	Destination         string    `json:"destination"`
	RollCall            time.Time `json:"rollCall"`
	UnknownRollCallDate bool      `json:"unknownRollCallDate"` //boolean indicates if RollCall date is unknown. False value may still indicate time of day was found.
	SeatCount           int       `json:"seatCount"`
	SeatType            string    `json:"seatType"`
	Cancelled           bool      `json:"cancelled"`
	PhotoSource         string    `json:"photoSource"` //FB node id
	SourceDate          time.Time `json:"sourceDate"`  //FB node created time
}

//Representation of Photo Report by user
type PhotoReport struct {
	Location              string    
	PhotoSource         string    
	Comment            string 
	SubmitDate time.Time      
	IPAddress           string
}

/*
 * Representation of web server data
 */

type SAResponse struct {
	Status    int      `json:"status"`
	Error     string   `json:"error"`
	User      string   `json:"user"`
	AuthToken string   `json:"authToken"`
	Flights   []Flight `json:"flights"`
	Locations []string `json:"locations"`
	Terminals []Terminal `json:"terminals"`
	Data string	`json:"data"`
}
