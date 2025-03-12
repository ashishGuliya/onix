package protocol

import "time"

// Ack represents the acknowledgement sent in response to an API call.
type Ack struct {
	Status string     `json:"status,omitempty" validate:"required,oneof=ACK NACK,omitempty"` // Must be either ACK or NACK
	Tags   []TagGroup `json:"tags,omitempty"`                                                // Optional list of tags
}

// TagGroup represents a collection of tag objects with group-level attributes.
type TagGroup struct {
	Display    bool       `json:"display,omitempty"`    // If false, the group will not be displayed.
	Descriptor Descriptor `json:"descriptor,omitempty"` // Description of the TagGroup.
	List       []Tag      `json:"list,omitempty"`       // List of Tag objects in this group.
}

// Descriptor represents the physical description of something.
type Descriptor struct {
	Name           string         `json:"name,omitempty"`
	Code           string         `json:"code,omitempty"`
	ShortDesc      string         `json:"short_desc,omitempty"`
	LongDesc       string         `json:"long_desc,omitempty"`
	AdditionalDesc AdditionalDesc `json:"additional_desc,omitempty"`
	Media          []MediaFile    `json:"media,omitempty"`
	Images         []Image        `json:"images,omitempty"`
}

// AdditionalDesc represents additional descriptive information.
type AdditionalDesc struct {
	URL         string `json:"url,omitempty"`
	ContentType string `json:"content_type,omitempty" validate:"oneof=text/plain text/html application/json,omitempty"`
}

// Tag represents a tag containing extended metadata.
type Tag struct {
	Descriptor Descriptor `json:"descriptor,omitempty"` // Description of the tag.
	Value      string     `json:"value,omitempty"`      // The value of the tag, set by the BPP.
	Display    bool       `json:"display,omitempty"`    // Indicates if the tag should be displayed.
}

// MediaFile contains a URL to a media file along with metadata.
type MediaFile struct {
	MIMEType  string `json:"mimetype,omitempty"`                     // Nature and format of the file (RFC 6838)
	URL       string `json:"url,omitempty" validate:"uri,omitempty"` // URL of the file
	Signature string `json:"signature,omitempty"`                    // Digital signature of the file
	DSA       string `json:"dsa,omitempty"`                          // Signing algorithm used by the sender
}

// Image represents an image with its metadata.
type Image struct {
	URL      string `json:"url,omitempty" validate:"uri,omitempty"`                               // URL to the image (can be remote or data URL)
	SizeType string `json:"size_type,omitempty" validate:"oneof=xs sm md lg xl custom,omitempty"` // Size type of the image
	Width    string `json:"width,omitempty"`                                                      // Width in pixels
	Height   string `json:"height,omitempty"`                                                     // Height in pixels
}

type Error struct {
	Code    string `json:"code,omitempty"`
	Paths   string `json:"paths,omitempty"`
	Message string `json:"message,omitempty"`
}

// Message struct (Contains either Ack or Error)
type Message struct {
	Ack *Ack `json:"ack,omitempty"`
}

type MessageForSearch struct {
	Intent Intent `json:"intent,omitempty"`
}

type MessageForOnSearch struct {
	Catalog Catalog
}

type OnSearchRequest struct {
	Context Context            `json:"context,omitempty"`
	Message MessageForOnSearch `json:"message,omitempty"`
}

type Response struct {
	Message Message `json:"message,omitempty"`
	Error   *Error  `json:"error,omitempty"`
}

type SearchRequest struct {
	Context Context          `json:"context,omitempty"`
	Message MessageForSearch `json:"message,omitempty"`
}

// Context represents the context object in the Beckn protocol.
type Context struct {
	Domain        string    `json:"domain,omitempty"`
	Country       string    `json:"country,omitempty"`
	City          string    `json:"city,omitempty"`
	Action        string    `json:"action,omitempty"`
	CoreVersion   string    `json:"core_version,omitempty"`
	BapID         string    `json:"bap_id,omitempty"`
	BapURI        string    `json:"bap_uri,omitempty"`
	BppID         string    `json:"bpp_id,omitempty"`
	BppURI        string    `json:"bpp_uri,omitempty"`
	TransactionID string    `json:"transaction_id,omitempty"`
	MessageID     string    `json:"message_id,omitempty"`
	Timestamp     time.Time `json:"timestamp,omitempty"`
}

type Location struct {
	ID          string     `json:"id,omitempty"`         // Unique identifier for the location
	Descriptor  Descriptor `json:"descriptor,omitempty"` // Description of the location
	MapURL      string     `json:"map_url,omitempty"`    // URL to the map of the location
	GPS         GPS        `json:"gps,omitempty"`        // GPS coordinates of the location
	Address     Address    `json:"address,omitempty"`    // Address details of the location
	City        City       `json:"city,omitempty"`       // City where the location is situated
	District    string     `json:"district,omitempty"`   // District where the location is situated
	State       State      `json:"state,omitempty"`      // State where the location is situated
	Country     Country    `json:"country,omitempty"`    // Country where the location is situated
	AreaCode    string     `json:"area_code,omitempty"`  // Area code of the location
	Circle      Circle     `json:"circle,omitempty"`     // Circular region defining the location
	Polygon     string     `json:"polygon,omitempty"`    // Boundary polygon of the location
	ThreeDSpace string     `json:"3dspace,omitempty"`    // 3D spatial region of the location
	Rating      float64    `json:"rating,omitempty"`     // Rating value of the location
}

type GPS string

// Address represents a postal address as a string.
type Address string

// Country represents a country with its name and ISO code.
type Country struct {
	Name string `json:"name,omitempty"` // Name of the country
	Code string `json:"code,omitempty"` // Country code as per ISO 3166-1 and ISO 3166-2
}

// City represents a city with a name and a code.
type City struct {
	Name string `json:"name,omitempty"` // Name of the city
	Code string `json:"code,omitempty"` // City code
}

// State represents a geopolitical region inside a country.
type State struct {
	Name string `json:"name,omitempty"` // Name of the state
	Code string `json:"code,omitempty"` // State code as per country or international standards
}

// Circle represents a circular region defined by a GPS coordinate and a radius.
type Circle struct {
	GPS    GPS    `json:"gps,omitempty"`    // Center GPS coordinate
	Radius string `json:"radius,omitempty"` // Radius (Scalar value)
}

type Intent struct {
	Descriptor  Descriptor  `json:"descriptor,omitempty"`  // Free text search strings, raw audio, etc.
	Provider    Provider    `json:"provider,omitempty"`    // Provider from which the customer wants to place the order
	Fulfillment Fulfillment `json:"fulfillment,omitempty"` // Details on how the customer wants their order fulfilled
	Payment     Payment     `json:"payment,omitempty"`     // Payment details
	Category    Category    `json:"category,omitempty"`    // Item category details
	Offer       Offer       `json:"offer,omitempty"`       // Details of the offer the customer wants to avail
	Item        Item        `json:"item,omitempty"`        // Details of the item that the consumer wants to order
	Tags        []TagGroup  `json:"tags,omitempty"`        // List of tag groups
}

type Duration string

type ItemQuantity struct {
	Allocated QuantityDetail `json:"allocated,omitempty"`
	Available QuantityDetail `json:"available,omitempty"`
	Maximum   QuantityDetail `json:"maximum,omitempty"`
	Minimum   QuantityDetail `json:"minimum,omitempty"`
	Selected  QuantityDetail `json:"selected,omitempty"`
	Unitized  QuantityDetail `json:"unitized,omitempty"`
}

type QuantityDetail struct {
	Count   int    `json:"count,omitempty"`
	Measure Scalar `json:"measure,omitempty"`
}

type DecimalValue string

type Scalar struct {
	Type           string       `json:"type,omitempty"`
	Value          DecimalValue `json:"value,omitempty"`
	EstimatedValue DecimalValue `json:"estimated_value,omitempty"`
	ComputedValue  DecimalValue `json:"computed_value,omitempty"`
	Range          ScalarRange  `json:"range,omitempty"`
	Unit           string       `json:"unit,omitempty"`
}

type ScalarRange struct {
	Min DecimalValue `json:"min,omitempty"`
	Max DecimalValue `json:"max,omitempty"`
}

type Organization struct {
	Descriptor Descriptor `json:"descriptor,omitempty"`
	Address    Address    `json:"address,omitempty"`
	State      State      `json:"state,omitempty"`
	City       City       `json:"city,omitempty"`
	Contact    Contact    `json:"contact,omitempty"`
}

type JCardProperty struct {
	Name       string                 `json:"name,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	ValueType  string                 `json:"value_type,omitempty"`
	Value      interface{}            `json:"value,omitempty"`
}

type JCard struct {
	CardType   string          `json:"card_type,omitempty"`
	Properties []JCardProperty `json:"properties,omitempty"`
}

type Contact struct {
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
	JCard JCard  `json:"jcard,omitempty"`
}

type Price struct {
	Currency       string `json:"currency,omitempty"`
	Value          string `json:"value,omitempty"`
	EstimatedValue string `json:"estimated_value,omitempty"`
	ComputedValue  string `json:"computed_value,omitempty"`
	ListedValue    string `json:"listed_value,omitempty"`
	OfferedValue   string `json:"offered_value,omitempty"`
	MinimumValue   string `json:"minimum_value,omitempty"`
	MaximumValue   string `json:"maximum_value,omitempty"`
}

type Option struct {
	ID         string     `json:"id,omitempty"`
	Descriptor Descriptor `json:"descriptor,omitempty"`
}

type CancellationEvent struct {
	Time                  string     `json:"time,omitempty"`         // ISO 8601 date-time format
	CancelledBy           string     `json:"cancelled_by,omitempty"` // "CONSUMER" or "PROVIDER"
	Reason                Option     `json:"reason,omitempty"`
	AdditionalDescription Descriptor `json:"additional_description,omitempty"`
}

type XInput struct {
	Form     Form `json:"form,omitempty"`
	Required bool `json:"required,omitempty"`
}

type Form struct {
	URL          string            `json:"url,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
	MimeType     string            `json:"mime_type,omitempty"`
	SubmissionID string            `json:"submission_id,omitempty"`
}

type Fee struct {
	Percentage string `json:"percentage,omitempty"`
	Amount     Price  `json:"amount,omitempty"`
}

type CancellationTerm struct {
	FulfillmentState FulfillmentState `json:"fulfillment_state,omitempty"`
	ReasonRequired   bool             `json:"reason_required,omitempty"`
	CancelBy         Time             `json:"cancel_by,omitempty"`
	CancellationFee  Fee              `json:"cancellation_fee,omitempty"`
	XInput           XInput           `json:"xinput,omitempty"`
	ExternalRef      MediaFile        `json:"external_ref,omitempty"`
}

type ReplacementTerm struct {
	FulfillmentState *State     `json:"fulfillment_state,omitempty"`
	ReplaceWithin    *Time      `json:"replace_within,omitempty"`
	ExternalRef      *MediaFile `json:"external_ref,omitempty"`
}

type ReturnTerm struct {
	FulfillmentState     *State    `json:"fulfillment_state,omitempty"`
	ReturnEligible       bool      `json:"return_eligible,omitempty"`
	ReturnTime           *Time     `json:"return_time,omitempty"`
	ReturnLocation       *Location `json:"return_location,omitempty"`
	FulfillmentManagedBy string    `json:"fulfillment_managed_by,omitempty"`
}

type RefundTerm struct {
	FulfillmentState *State `json:"fulfillment_state,omitempty"`
	RefundEligible   bool   `json:"refund_eligible,omitempty"`
	RefundWithin     *Time  `json:"refund_within,omitempty"`
	RefundAmount     *Price `json:"refund_amount,omitempty"`
}

type Item struct {
	ID                 string             `json:"id,omitempty"`
	ParentItemID       string             `json:"parent_item_id,omitempty"`
	ParentItemQuantity ItemQuantity       `json:"parent_item_quantity,omitempty"`
	Descriptor         Descriptor         `json:"descriptor,omitempty"`
	Creator            Organization       `json:"creator,omitempty"`
	Price              Price              `json:"price,omitempty"`
	Quantity           ItemQuantity       `json:"quantity,omitempty"`
	CategoryIDs        []string           `json:"category_ids,omitempty"`
	FulfillmentIDs     []string           `json:"fulfillment_ids,omitempty"`
	LocationIDs        []string           `json:"location_ids,omitempty"`
	PaymentIDs         []string           `json:"payment_ids,omitempty"`
	AddOns             []AddOn            `json:"add_ons,omitempty"`
	CancellationTerms  []CancellationTerm `json:"cancellation_terms,omitempty"`
	RefundTerms        []RefundTerm       `json:"refund_terms,omitempty"`
	ReplacementTerms   []ReplacementTerm  `json:"replacement_terms,omitempty"`
	ReturnTerms        []ReturnTerm       `json:"return_terms,omitempty"`
	XInput             XInput             `json:"xinput,omitempty"`
	Time               Time               `json:"time,omitempty"`
	Rateable           bool               `json:"rateable,omitempty"`
	Rating             float64            `json:"rating,omitempty"`
	Matched            bool               `json:"matched,omitempty"`
	Related            bool               `json:"related,omitempty"`
	Recommended        bool               `json:"recommended,omitempty"`
	TTL                string             `json:"ttl,omitempty"`
	Tags               []TagGroup         `json:"tags,omitempty"`
}

type AddOn struct {
	ID         string     `json:"id,omitempty"`
	Descriptor Descriptor `json:"descriptor,omitempty"`
	Price      Price      `json:"price,omitempty"`
}

type Category struct {
	ID               string     `json:"id,omitempty"`
	ParentCategoryID string     `json:"parent_category_id,omitempty"`
	Descriptor       Descriptor `json:"descriptor,omitempty"`
	Time             Time       `json:"time,omitempty"`
	TTL              string     `json:"ttl,omitempty"`
	Tags             []TagGroup `json:"tags,omitempty"`
}

type Customer struct {
	Person  *Person  `json:"person,omitempty"`
	Contact *Contact `json:"contact,omitempty"`
}

type Agent struct {
	Person       *Person       `json:"person,omitempty"`
	Contact      *Contact      `json:"contact,omitempty"`
	Organization *Organization `json:"organization,omitempty"`
	Rating       *float64      `json:"rating,omitempty"`
}

type Stop struct {
	ID            string         `json:"id,omitempty"`
	ParentStopID  string         `json:"parent_stop_id,omitempty"`
	Location      *Location      `json:"location,omitempty"`
	Type          string         `json:"type,omitempty"`
	Time          *Time          `json:"time,omitempty"`
	Instructions  *Descriptor    `json:"instructions,omitempty"`
	Contact       *Contact       `json:"contact,omitempty"`
	Person        *Person        `json:"person,omitempty"`
	Authorization *Authorization `json:"authorization,omitempty"`
}

type Authorization struct {
	Type      string    `json:"type,omitempty"`
	Token     string    `json:"token,omitempty"`
	ValidFrom time.Time `json:"valid_from,omitempty"`
	ValidTo   time.Time `json:"valid_to,omitempty"`
	Status    string    `json:"status,omitempty"`
}

type Credential struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type" default:"VerifiableCredential,omitempty"`
	URL  string `json:"url,omitempty"`
}

type Language struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type Skill struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type Person struct {
	ID        string       `json:"id,omitempty"`
	URL       string       `json:"url,omitempty"`
	Name      string       `json:"name,omitempty"`
	Image     *Image       `json:"image,omitempty"`
	Age       *Duration    `json:"age,omitempty"`
	DOB       string       `json:"dob,omitempty"` // Format: "date"
	Gender    string       `json:"gender,omitempty"`
	Creds     []Credential `json:"creds,omitempty"`
	Languages []Language   `json:"languages,omitempty"`
	Skills    []Skill      `json:"skills,omitempty"`
	Tags      []TagGroup   `json:"tags,omitempty"`
}

type Vehicle struct {
	Category         string `json:"category,omitempty"`
	Capacity         int    `json:"capacity,omitempty"`
	Make             string `json:"make,omitempty"`
	Model            string `json:"model,omitempty"`
	Size             string `json:"size,omitempty"`
	Variant          string `json:"variant,omitempty"`
	Color            string `json:"color,omitempty"`
	EnergyType       string `json:"energy_type,omitempty"`
	Registration     string `json:"registration,omitempty"`
	WheelsCount      string `json:"wheels_count,omitempty"`
	CargoVolume      string `json:"cargo_volume,omitempty"`
	WheelchairAccess string `json:"wheelchair_access,omitempty"`
	Code             string `json:"code,omitempty"`
	EmissionStandard string `json:"emission_standard,omitempty"`
}
type Fulfillment struct {
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type,omitempty"`
	Rateable bool             `json:"rateable,omitempty"`
	Rating   float64          `json:"rating,omitempty"`
	State    FulfillmentState `json:"state,omitempty"`
	Tracking bool             `json:"tracking,omitempty"`
	Customer Customer         `json:"customer,omitempty"`
	Agent    Agent            `json:"agent,omitempty"`
	Contact  Contact          `json:"contact,omitempty"`
	Vehicle  Vehicle          `json:"vehicle,omitempty"`
	Stops    []Stop           `json:"stops,omitempty"`
	Path     string           `json:"path,omitempty"`
	Tags     []TagGroup       `json:"tags,omitempty"`
}

type FulfillmentState struct {
	Descriptor Descriptor `json:"descriptor,omitempty"`
	UpdatedAt  string     `json:"updated_at,omitempty"`
	UpdatedBy  string     `json:"updated_by,omitempty"`
}

type Payment struct {
	ID          string        `json:"id,omitempty"`
	CollectedBy string        `json:"collected_by,omitempty"` //Enum: "bap" or "bpp"
	URL         string        `json:"url,omitempty"`
	Params      PaymentParams `json:"params,omitempty"`
	Type        string        `json:"type,omitempty"`
	Status      string        `json:"status,omitempty"`
	Time        Time          `json:"time,omitempty"`
	Tags        []TagGroup    `json:"tags,omitempty"`
}

type PaymentParams struct {
	TransactionID               string `json:"transaction_id,omitempty"`
	Amount                      string `json:"amount,omitempty"`
	Currency                    string `json:"currency,omitempty"`
	BankCode                    string `json:"bank_code,omitempty"`
	BankAccountNumber           string `json:"bank_account_number,omitempty"`
	VirtualPaymentAddress       string `json:"virtual_payment_address,omitempty"`
	SourceBankCode              string `json:"source_bank_code,omitempty"`
	SourceBankAccountNumber     string `json:"source_bank_account_number,omitempty"`
	SourceVirtualPaymentAddress string `json:"source_virtual_payment_address,omitempty"`
}

type Provider struct {
	ID           string        `json:"id,omitempty"`
	Descriptor   Descriptor    `json:"descriptor,omitempty"`
	CategoryID   string        `json:"category_id,omitempty"`
	Rating       float64       `json:"rating,omitempty"`
	Time         Time          `json:"time,omitempty"`
	Categories   []Category    `json:"categories,omitempty"`
	Fulfillments []Fulfillment `json:"fulfillments,omitempty"`
	Payments     []Payment     `json:"payments,omitempty"`
	Locations    []Location    `json:"locations,omitempty"`
	Offers       []Offer       `json:"offers,omitempty"`
	Items        []Item        `json:"items,omitempty"`
	Exp          string        `json:"exp,omitempty"`
	Rateable     bool          `json:"rateable,omitempty"`
	TTL          int           `json:"ttl,omitempty"`
	Tags         []TagGroup    `json:"tags,omitempty"`
}

type Time struct {
	Label     string    `json:"label,omitempty"`
	Timestamp string    `json:"timestamp,omitempty"`
	Duration  Duration  `json:"duration,omitempty"`
	Range     TimeRange `json:"range,omitempty"`
	Days      string    `json:"days,omitempty"`
	Schedule  Schedule  `json:"schedule,omitempty"`
}

type Schedule struct {
	Frequency *Duration   `json:"frequency,omitempty"` // Duration of the recurring event
	Holidays  []time.Time `json:"holidays,omitempty"`  // Dates when the event won't occur
	Times     []time.Time `json:"times,omitempty"`     // Specific times when the event occurs
}

type TimeRange struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

type Offer struct {
	ID          string     `json:"id,omitempty"`
	Descriptor  Descriptor `json:"descriptor,omitempty"`
	LocationIDs []string   `json:"location_ids,omitempty"`
	CategoryIDs []string   `json:"category_ids,omitempty"`
	ItemIDs     []string   `json:"item_ids,omitempty"`
	Time        Time       `json:"time,omitempty"`
	Tags        []TagGroup `json:"tags,omitempty"`
}

type Catalog struct {
	Descriptor   Descriptor    `json:"bpp/descriptor,omitempty"`
	Fulfillments []Fulfillment `json:"bpp/fulfillments,omitempty"`
	Payments     []Payment     `json:"payments,omitempty"`
	Offers       []Offer       `json:"offers,omitempty"`
	Providers    []Provider    `json:"bpp/providers,omitempty"`
	Exp          time.Time     `json:"exp,omitempty"`
	TTL          string        `json:"ttl,omitempty"`
}
