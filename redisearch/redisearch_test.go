package redisearch

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createClient(indexName string) *Client {
	value, exists := os.LookupEnv("REDISEARCH_TEST_HOST")
	host := "localhost:6379"
	if exists && value != "" {
		host = value
	}
	return NewClient(host, indexName)
}

func createAutocompleter(indexName string) *Autocompleter {
	value, exists := os.LookupEnv("REDISEARCH_TEST_HOST")
	host := "localhost:6379"
	if exists && value != "" {
		host = value
	}
	return NewAutocompleter(host, indexName)
}

func TestClient(t *testing.T) {

	c := createClient("testing")
	defer c.Close()

	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("foo"))
	c.Drop()
	if err := c.CreateIndex(sc); err != nil {
		t.Fatal(err)
	}

	docs := make([]Document, 100)
	for i := 0; i < 100; i++ {
		docs[i] = NewDocument(fmt.Sprintf("doc%d", i), float32(i)/float32(100)).Set("foo", "hello world")
	}

	if err := c.IndexOptions(DefaultIndexingOptions, docs...); err != nil {
		t.Fatal(err)
	}

	// Test it again
	if err := c.IndexOptions(DefaultIndexingOptions, docs...); err == nil {
		t.Fatal("Expected error for duplicate document")
	} else {
		if merr, ok := err.(MultiError); !ok {
			t.Fatal("error not a multi error")
		} else {
			assert.Equal(t, 100, len(merr))
			assert.NotEmpty(t, merr)
			//fmt.Println("Got errors: ", merr)
		}
	}

	docs, total, err := c.Search(NewQuery("hello world"))

	assert.Nil(t, err)
	assert.Equal(t, 100, total)
	assert.Equal(t, 10, len(docs))

	fmt.Println(docs, total, err)
}

func TestEscaping(t *testing.T) {
	tests := []string{
		",",
		".",
		"<",
		">",
		"{",
		"}",
		"[",
		"]",
		"\"",
		"'",
		":",
		";",
		"!",
		"@",
		"#",
		"$",
		"%",
		"^",
		"&",
		"*",
		"(",
		")",
		"-",
		"+",
		"=",
		"~",
		"\\",
	}
	for _, s := range tests {
		fmt.Println(escapeString(s))
	}
}

func TestInfo(t *testing.T) {
	c := createClient("testung")
	defer c.Close()

	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("foo")).
		AddField(NewSortableNumericField("bar"))
	c.Drop()
	assert.Nil(t, c.CreateIndex(sc))

	info, err := c.Info()
	assert.Nil(t, err)
	fmt.Printf("%v\n", info)
}

func TestNumeric(t *testing.T) {
	c := createClient("testung")
	defer c.Close()

	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("foo")).
		AddField(NewSortableNumericField("bar"))
	c.Drop()
	assert.Nil(t, c.CreateIndex(sc))

	docs := make([]Document, 100)
	for i := 0; i < 100; i++ {
		docs[i] = NewDocument(fmt.Sprintf("doc%d", i), 1).Set("foo", "hello world").Set("bar", i)
	}

	assert.Nil(t, c.Index(docs...))

	docs, total, err := c.Search(NewQuery("hello world @bar:[50 100]").SetFlags(QueryNoContent | QueryWithScores))
	assert.Nil(t, err)
	assert.Equal(t, 10, len(docs))
	assert.Equal(t, 50, total)

	docs, total, err = c.Search(NewQuery("hello world @bar:[40 90]").SetSortBy("bar", false))
	assert.Nil(t, err)
	assert.Equal(t, 10, len(docs))
	assert.Equal(t, 51, total)
	assert.Equal(t, "doc90", docs[0].Id)
	assert.Equal(t, "doc89", docs[1].Id)
	assert.Equal(t, "doc81", docs[9].Id)

	docs, total, err = c.Search(NewQuery("hello world @bar:[40 90]").
		SetSortBy("bar", true).
		SetReturnFields("foo"))
	assert.Nil(t, err)
	assert.Equal(t, 10, len(docs))
	assert.Equal(t, 51, total)
	assert.Equal(t, "doc40", docs[0].Id)
	assert.Equal(t, "hello world", docs[0].Properties["foo"])
	assert.Nil(t, docs[0].Properties["bar"])
	assert.Equal(t, "doc41", docs[1].Id)
	assert.Equal(t, "doc49", docs[9].Id)
	fmt.Println(docs)

	// Try "Explain"
	explain, err := c.Explain(NewQuery("hello world @bar:[40 90]"))
	assert.Nil(t, err)
	assert.NotNil(t, explain)
	fmt.Println(explain)
}

func TestNoIndex(t *testing.T) {
	c := createClient("testung")
	defer c.Close()
	c.Drop()

	sc := NewSchema(DefaultOptions).
		AddField(NewTextFieldOptions("f1", TextFieldOptions{Sortable: true, NoIndex: true, Weight: 1.0})).
		AddField(NewTextField("f2"))

	err := c.CreateIndex(sc)
	assert.Nil(t, err)

	props := make(map[string]interface{})
	props["f1"] = "MarkZZ"
	props["f2"] = "MarkZZ"

	err = c.Index(Document{Id: "doc1", Properties: props})
	assert.Nil(t, err)

	props["f1"] = "MarkAA"
	props["f2"] = "MarkAA"
	err = c.Index(Document{Id: "doc2", Properties: props})
	assert.Nil(t, err)

	_, total, err := c.Search(NewQuery("@f1:Mark*"))
	assert.Nil(t, err)
	assert.Equal(t, 0, total)

	_, total, err = c.Search(NewQuery("@f2:Mark*"))
	assert.Equal(t, 2, total)

	docs, total, err := c.Search(NewQuery("@f2:Mark*").SetSortBy("f1", false))
	assert.Equal(t, 2, total)
	assert.Equal(t, "doc1", docs[0].Id)

	docs, total, err = c.Search(NewQuery("@f2:Mark*").SetSortBy("f1", true))
	assert.Equal(t, 2, total)
	assert.Equal(t, "doc2", docs[0].Id)
}

func TestHighlight(t *testing.T) {
	c := createClient("testung")
	defer c.Close()

	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("foo")).
		AddField(NewTextField("bar"))
	c.Drop()
	assert.Nil(t, c.CreateIndex(sc))

	docs := make([]Document, 100)
	for i := 0; i < 100; i++ {
		docs[i] = NewDocument(fmt.Sprintf("doc%d", i), 1).Set("foo", "hello world").Set("bar", "hello world foo bar baz")
	}
	c.Index(docs...)

	q := NewQuery("hello").Highlight([]string{"foo"}, "[", "]")
	docs, _, err := c.Search(q)
	assert.Nil(t, err)

	assert.Equal(t, 10, len(docs))
	for _, d := range docs {
		assert.Equal(t, "[hello] world", d.Properties["foo"])
		assert.Equal(t, "hello world foo bar baz", d.Properties["bar"])
	}

	q = NewQuery("hello world baz").Highlight([]string{"foo", "bar"}, "{", "}")
	docs, _, err = c.Search(q)
	assert.Nil(t, err)

	assert.Equal(t, 10, len(docs))
	for _, d := range docs {
		assert.Equal(t, "{hello} {world}", d.Properties["foo"])
		assert.Equal(t, "{hello} {world} foo bar {baz}", d.Properties["bar"])
	}

	// test RETURN contradicting HIGHLIGHT
	q = NewQuery("hello").Highlight([]string{"foo"}, "[", "]").SetReturnFields("bar")
	docs, _, err = c.Search(q)
	assert.Nil(t, err)

	assert.Equal(t, 10, len(docs))
	for _, d := range docs {
		assert.Equal(t, nil, d.Properties["foo"])
		assert.Equal(t, "hello world foo bar baz", d.Properties["bar"])
	}

	c.Drop()
}

func TestSammurize(t *testing.T) {
	c := createClient("testung")
	defer c.Close()

	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("foo")).
		AddField(NewTextField("bar"))
	c.Drop()
	assert.Nil(t, c.CreateIndex(sc))

	docs := make([]Document, 10)
	for i := 0; i < 10; i++ {
		docs[i] = NewDocument(fmt.Sprintf("doc%d", i), 1).
			Set("foo", "There are two sub-commands commands used for highlighting. One is HIGHLIGHT which surrounds matching text with an open and/or close tag; and the other is SUMMARIZE which splits a field into contextual fragments surrounding the found terms. It is possible to summarize a field, highlight a field, or perform both actions in the same query.").Set("bar", "hello world foo bar baz")
	}
	c.Index(docs...)

	q := NewQuery("commands fragments fields").Summarize("foo")
	docs, _, err := c.Search(q)
	assert.Nil(t, err)

	assert.Equal(t, 10, len(docs))
	for _, d := range docs {
		assert.Equal(t, "are two sub-commands commands used for highlighting. One is HIGHLIGHT which surrounds... other is SUMMARIZE which splits a field into contextual fragments surrounding the found terms. It is possible to summarize a field, highlight a field, or perform both actions in the... ", d.Properties["foo"])
		assert.Equal(t, "hello world foo bar baz", d.Properties["bar"])
	}

	q = NewQuery("commands fragments fields").
		Highlight([]string{"foo"}, "[", "]").
		SummarizeOptions(SummaryOptions{
			Fields:       []string{"foo"},
			Separator:    "\r\n",
			FragmentLen:  10,
			NumFragments: 5},
		)
	docs, _, err = c.Search(q)
	assert.Nil(t, err)

	assert.Equal(t, 10, len(docs))
	for _, d := range docs {
		assert.Equal(t, "are two sub-[commands] [commands] used for highlighting. One is\r\na [field] into contextual [fragments] surrounding the found terms. It is possible to summarize a [field], highlight a [field], or\r\n", d.Properties["foo"])
		assert.Equal(t, "hello world foo bar baz", d.Properties["bar"])
	}
}

func TestTags(t *testing.T) {
	c := createClient("myIndex")
	defer c.Close()

	// Create a schema
	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("title")).
		AddField(NewTagFieldOptions("tags", TagFieldOptions{Separator: ';'})).
		AddField(NewTagField("tags2"))

	// Drop an existing index. If the index does not exist an error is returned
	c.Drop()

	// Create the index with the given schema
	if err := c.CreateIndex(sc); err != nil {
		log.Fatal(err)
	}

	// Create a document with an id and given score
	doc := NewDocument("doc1", 1.0)
	doc.Set("title", "Hello world").
		Set("tags", "foo bar;bar,baz;  hello world").
		Set("tags2", "foo bar;bar,baz;  hello world")

	// Index the document. The API accepts multiple documents at a time
	if err := c.IndexOptions(DefaultIndexingOptions, doc); err != nil {
		log.Fatal(err)
	}

	assertNumResults := func(qs string, tagFilters []tagFilter, n int) {
		// Searching with tag filters
		q := NewQuery(qs)
		q.tagFilters = tagFilters
		q.SetFlags(QueryWithScores | QueryWithPayloads)
		_, total, err := c.Search(q)
		assert.Nil(t, err)

		assert.Equal(t, n, total)
	}

	assertNumResults("", []tagFilter{
		{
			field:  "tags",
			values: []string{"bar"},
		},
	}, 0)
	assertNumResults("", []tagFilter{
		{
			field:  "tags",
			values: []string{"foo bar"},
		},
	}, 1)
	assertNumResults("", []tagFilter{
		{
			field:  "tags",
			values: []string{"foo bar", "bar"},
		},
	}, 1)
	assertNumResults("", []tagFilter{
		{
			field:  "tags",
			values: []string{"bar,baz"},
		},
	}, 1)
	assertNumResults("", []tagFilter{
		{
			field:  "tags",
			values: []string{"hello world"},
		},
	}, 1)
	assertNumResults("", []tagFilter{
		{
			field:  "tags",
			values: []string{"hello world"},
		},
	}, 1)
	assertNumResults("", []tagFilter{
		{
			field:  "tags",
			values: []string{"hello world"},
		},
		{
			field:  "tags2",
			values: []string{"foo bar\\;bar"}, // for the sake of simplicity, disallow tag filters other than comma
		},
	}, 1)
	assertNumResults("", []tagFilter{
		{
			field:  "tags",
			values: []string{"hello world"},
		},
		{
			field:  "tags2",
			values: []string{"foo bar\\;bar"}, // for the sake of simplicity, disallow tag filters other than comma
		},
	}, 1)
	assertNumResults("foo bar", nil, 0)
	assertNumResults("hello world", nil, 1)
}

func TestInFields(t *testing.T) {
	c := createClient("myIndex")
	defer c.Close()

	// Create a schema
	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("title")).
		AddField(NewTextField("description")).
		AddField(NewTextField("keywords"))

	// Drop an existing index. If the index does not exist an error is returned
	c.Drop()

	// Create the index with the given schema
	if err := c.CreateIndex(sc); err != nil {
		log.Fatal(err)
	}

	// Create a document with an id and given score
	doc := NewDocument("doc1", 1.0)
	doc.Set("title", "Hello world").
		Set("description", "foo bar").
		Set("keyword", "iphone")

	// Index the document. The API accepts multiple documents at a time
	if err := c.IndexOptions(DefaultIndexingOptions, doc); err != nil {
		log.Fatal(err)
	}

	assertNumResults := func(qs string, fields []string, n int) {
		// Searching with tag filters
		q := NewQuery(qs).SetInFields(fields...)
		q.SetFlags(QueryWithScores | QueryWithPayloads)
		_, total, err := c.Search(q)
		assert.Nil(t, err)

		assert.Equal(t, n, total)
	}

	assertNumResults("hello world", []string{"title"}, 1)
	assertNumResults("hello world", []string{"description", "keyword"}, 0)
	assertNumResults("foo bar", []string{"description", "keyword"}, 1)
}

func TestNumericFilters(t *testing.T) {
	c := createClient("myIndex")
	defer c.Close()

	// Create a schema
	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("title")).
		AddField(NewSortableNumericField("start")).
		AddField(NewNumericField("end"))

	// Drop an existing index. If the index does not exist an error is returned
	c.Drop()

	// Create the index with the given schema
	if err := c.CreateIndex(sc); err != nil {
		log.Fatal(err)
	}

	// Create a document with an id and given score
	doc := NewDocument("doc1", 1.0).
		Set("title", "foo").
		Set("start", 1).
		Set("end", 2)

	// Index the document. The API accepts multiple documents at a time
	if err := c.IndexOptions(DefaultIndexingOptions, doc); err != nil {
		log.Fatal(err)
	}

	assertNumResults := func(predicate Predicate, n int) {
		// Searching with tag filters
		q := NewQuery("").AddPredicate(predicate)
		q.SetFlags(QueryWithScores | QueryWithPayloads)
		_, total, err := c.Search(q)
		assert.Nil(t, err)

		assert.Equal(t, n, total)
	}
	assertNumResults(InRange("start", 0, 1, false, false), 0)
	assertNumResults(InRange("start", 0, 1, false, true), 1)
	assertNumResults(LessThan("start", 1), 0)
	assertNumResults(LessThanEquals("start", 1), 1)
	assertNumResults(InRange("end", 2, 3, false, false), 0)
	assertNumResults(InRange("end", 2, 3, true, false), 1)
	assertNumResults(GreaterThan("end", 2), 0)
	assertNumResults(GreaterThanEquals("end", 2), 1)
	assertNumResults(Equals("start", 1), 1)
	assertNumResults(Equals("start", 2), 0)
}

func TestSuggest(t *testing.T) {

	a := createAutocompleter("testing")

	// Add Terms to the Autocompleter
	terms := make([]Suggestion, 10)
	for i := 0; i < 10; i++ {
		terms[i] = Suggestion{Term: fmt.Sprintf("foo %d", i),
			Score: 1.0, Payload: fmt.Sprintf("bar %d", i)}
	}
	err := a.AddTerms(terms...)
	assert.Nil(t, err)

	// Retrieve Terms From Autocompleter - Without Payloads / Scores
	suggestions, err := a.SuggestOpts("f", SuggestOptions{Num: 10})
	assert.Nil(t, err)
	assert.Equal(t, 10, len(suggestions))
	for _, suggestion := range suggestions {
		assert.Contains(t, suggestion.Term, "foo")
		assert.Equal(t, suggestion.Payload, "")
		assert.Zero(t, suggestion.Score)
	}

	// Retrieve Terms From Autocompleter - With Payloads & Scores
	suggestions, err = a.SuggestOpts("f", SuggestOptions{Num: 10, WithScores: true, WithPayloads: true})
	assert.Nil(t, err)
	assert.Equal(t, 10, len(suggestions))
	for _, suggestion := range suggestions {
		assert.Contains(t, suggestion.Term, "foo")
		assert.Contains(t, suggestion.Payload, "bar")
		assert.NotZero(t, suggestion.Score)
	}
}

func TestDelete(t *testing.T) {
	c := createClient("testung")
	defer c.Close()

	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("foo"))

	err := c.Drop()
	assert.Nil(t, err)
	assert.Nil(t, c.CreateIndex(sc))

	var info *IndexInfo

	// validate that the index is empty
	info, err = c.Info()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), info.DocCount)

	doc := NewDocument("doc1", 1.0)
	doc.Set("foo", "Hello world")

	err = c.IndexOptions(DefaultIndexingOptions, doc)
	assert.Nil(t, err)

	// now we should have 1 document (id = doc1)
	info, err = c.Info()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), info.DocCount)

	// delete the document from the index
	err = c.Delete("doc1", true)
	assert.Nil(t, err)

	// validate that the index is empty again
	info, err = c.Info()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), info.DocCount)
}

func ExampleClient() {

	// Create a client. By default a client is schemaless
	// unless a schema is provided when creating the index
	c := createClient("myIndex")
	defer c.Close()

	// Create a schema
	sc := NewSchema(DefaultOptions).
		AddField(NewTextField("body")).
		AddField(NewTextFieldOptions("title", TextFieldOptions{Weight: 5.0, Sortable: true})).
		AddField(NewNumericField("date"))

	// Drop an existing index. If the index does not exist an error is returned
	c.Drop()

	// Create the index with the given schema
	if err := c.CreateIndex(sc); err != nil {
		log.Fatal(err)
	}

	// Create a document with an id and given score
	doc := NewDocument("doc1", 1.0)
	doc.Set("title", "Hello world").
		Set("body", "foo bar").
		Set("date", time.Now().Unix())

	// Index the document. The API accepts multiple documents at a time
	if err := c.IndexOptions(DefaultIndexingOptions, doc); err != nil {
		log.Fatal(err)
	}

	// Searching with limit and sorting
	docs, total, err := c.Search(NewQuery("hello world").
		Limit(0, 2).
		SetReturnFields("title"))

	fmt.Println(docs[0].Id, docs[0].Properties["title"], total, err)
	// Output: doc1 Hello world 1 <nil>
}
