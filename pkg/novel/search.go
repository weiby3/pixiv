package novel

import (
	"context"
	"net/url"
	"strconv"

	"github.com/NateScarlet/pixiv/pkg/client"
	"github.com/NateScarlet/pixiv/pkg/user"
	"github.com/tidwall/gjson"
)

// SearchResult holds search data and provide useful methods.
type SearchResult struct {
	JSON gjson.Result
}

// ForEach iterates through novel data items.
func (r SearchResult) ForEach(iterator func(key, value gjson.Result) bool) {
	r.JSON.Get("novel.data").ForEach(iterator)

}

// Novels appeared in the search result.
func (r SearchResult) Novels() []Novel {
	ret := make([]Novel, 0, int(r.JSON.Get("#").Int()))
	r.ForEach(func(key, value gjson.Result) bool {
		n := Novel{
			ID:          value.Get("id").String(),
			Title:       value.Get("title").String(),
			Description: value.Get("Description").String(),
			Author: user.User{
				ID:   value.Get("userId").String(),
				Name: value.Get("userName").String(),
			},
			TextCount:     value.Get("textCount").Int(),
			BookmarkCount: value.Get("bookmarkCount").Int(),
			Series: Series{
				ID:    value.Get("seriesId").String(),
				Title: value.Get("seriesTitle").String(),
			},
		}
		tagsData := value.Get("tags").Array()
		tags := make([]string, len(tagsData))
		for index, v := range tagsData {
			tags[index] = v.String()
		}
		n.Tags = tags

		ret = append(ret, n)
		return true
	})
	return ret

}

type Order string

const (
	OrderDateNull       Order = ""
	OrderDateAscending  Order = "date"
	OrderDateDescending Order = "date_d"
)

type Lang string

const (
	LangNull Lang = ""
	LangZH   Lang = "zh"
)

type WorkLang string

const (
	WorkLangNull WorkLang = ""
	WorkLangZHCN WorkLang = "zh-cn"
)

// SearchOptions for Search
type SearchOptions struct {
	Page     int
	Order    Order
	Lang     Lang
	WorkLang WorkLang
}

// SearchOption mutate SearchOptions
type SearchOption func(*SearchOptions)

// SearchOptionPage change page to retrive
func SearchOptionPage(page int) SearchOption {
	return func(so *SearchOptions) {
		so.Page = page
	}
}

func SearchOptionOrder(order Order) SearchOption {
	return func(so *SearchOptions) {
		so.Order = order
	}
}

func SearchOptionLang(lang Lang) SearchOption {
	return func(so *SearchOptions) {
		so.Lang = lang
	}
}

func SearchOptionWorkLang(workLang WorkLang) SearchOption {
	return func(so *SearchOptions) {
		so.WorkLang = workLang
	}
}

// Search calls pixiv novel search api.
func Search(ctx context.Context, query string, opts ...SearchOption) (result SearchResult, err error) {
	var args = new(SearchOptions)
	for _, i := range opts {
		i(args)
	}
	if args.Page < 1 {
		args.Page = 1
	}

	q := url.Values{}
	if args.Page != 1 {
		q.Set("p", strconv.Itoa(args.Page))
	}
	if args.Order != OrderDateNull {
		q.Set("order", string(args.Order))
	}
	if args.Lang != LangNull {
		q.Set("lang", string(args.Lang))
	}
	if args.WorkLang != WorkLangNull {
		q.Set("work_lang", string(args.WorkLang))
	}

	var c = client.For(ctx)
	resp, err := c.GetWithContext(ctx, c.EndpointURL("/ajax/search/novels/"+query, &q).String())
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := client.ParseAPIResult(resp.Body)
	if err != nil {
		return
	}
	result = SearchResult{JSON: body}
	return
}
