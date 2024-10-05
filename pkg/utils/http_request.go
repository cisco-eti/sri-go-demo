package utils

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/cisco-eti/sre-go-helloworld/pkg/models"
)

// HTTPRequest struct
type HTTPRequest struct {
	*http.Request
}

// GetPaginationLinks is a utility that helps paginated apis to return next page, previous page and last page links
func (req *HTTPRequest) GetPaginationLinks(responseSize int, limit int, offset int) models.Links {
	if !isPaginationParamsProper(responseSize, limit, offset) {
		return models.Links{}
	}
	path := req.URL
	values, _ := url.ParseQuery(path.RawQuery)
	values.Set("limit", strconv.Itoa(limit))
	values.Set("offset", "0")
	path.RawQuery = values.Encode()

	links := models.Links{
		First: path.String(),
	}

	nextOffset := offset + limit
	prevOffset := offset - limit

	if responseSize < limit {
		//no more pages so set last
		values.Set("offset", strconv.Itoa(offset))
		path.RawQuery = values.Encode()
		links.Last = path.String()
	} else {
		//likely more pages so set next
		values.Set("offset", strconv.Itoa(nextOffset))
		path.RawQuery = values.Encode()
		links.Next = path.String()
	}

	if prevOffset >= 0 {
		//previous page exists so set prev
		values.Set("offset", strconv.Itoa(prevOffset))
		path.RawQuery = values.Encode()
		links.Prev = path.String()
	}

	return links
}

func isPaginationParamsProper(responseSize int, limit int, offset int) bool {
	return responseSize >= 0 &&
		limit >= 0 &&
		offset >= 0
}
