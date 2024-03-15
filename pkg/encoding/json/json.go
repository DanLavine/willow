package json

const (
	ContentTypeJSON = "application/json"
)

type JsonEncoder struct{}

func (_ JsonEncoder) ContentType() string {
	return ContentTypeJSON
}
