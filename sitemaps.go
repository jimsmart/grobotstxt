package grobotstxt

type sitemapExtractor struct {
	sitemaps []string
}

func (f *sitemapExtractor) Sitemaps(robotsBody string) []string {
	Parse(robotsBody, f)
	return f.sitemaps
}

func Sitemaps(robotsBody string) []string {
	return (&sitemapExtractor{}).Sitemaps(robotsBody)
}

func (f *sitemapExtractor) HandleRobotsStart() {
	f.sitemaps = nil
}

func (f *sitemapExtractor) HandleSitemap(lineNum int, value string) {
	f.sitemaps = append(f.sitemaps, value)
}

func (f *sitemapExtractor) HandleRobotsEnd() {}

func (f *sitemapExtractor) HandleUserAgent(lineNum int, value string) {}

func (f *sitemapExtractor) HandleAllow(lineNum int, value string) {}

func (f *sitemapExtractor) HandleDisallow(lineNum int, value string) {}

func (f *sitemapExtractor) HandleUnknownAction(lineNum int, action, value string) {}
