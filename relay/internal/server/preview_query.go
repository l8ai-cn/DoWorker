package server

import "net/url"

func previewRawQuery(rawQuery string) (string, error) {
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", err
	}
	query.Del("token")
	return query.Encode(), nil
}
