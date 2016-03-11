This library provides a way to parse, scrape, and transform html5 pages using CSS selector queries.

The html5 parser wraps the exp/html parser at tip in Go's source repo.

Patches for defencies you find are greatly appreciated.

# installs just the html5 parser wrapper
go get go.marzhillstudios.com/pkg/go-html-transform/h5 

# installs the full html/transform package
go get go.marzhillstudios.com/pkg/go-html-transform/html/transform
You can see sample usage in the comments at the top of the page here: https://bitbucket.com/zaphar/source/browse/html/transform/transform.go

documentation for the library can be found here: https://godoc.org/go.marzhillstudios.com/pkg/go-html-transform/html/transform