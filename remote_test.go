package selenium

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

var grid = flag.Bool("test.grid", false, "skip tests that fail on Selenium Grid")
var executor = flag.String("test.executor", defaultExecutor, "executor URL")
var browserName = flag.String("test.browserName", "firefox", "browser to run tests on")

func init() {
	flag.BoolVar(&Trace, "trace", false, "trace HTTP requests and responses")
	flag.Parse()

	caps["browserName"] = *browserName
}

var caps Capabilities = make(Capabilities)

var runOnSauce *bool = flag.Bool("saucelabs", false, "run on sauce")

func newRemote(testName string, t *testing.T) (wd WebDriver) {
	var err error
	if wd, err = NewRemote(caps, *executor); err != nil {
		t.Fatalf("can't start session for test %s: %s", testName, err)
	}
	return wd
}

func TestStatus(t *testing.T) {
	if *grid {
		t.Skip()
	}
	t.Parallel()
	wd := newRemote("TestStatus", t)
	defer wd.Quit()

	status, err := wd.Status()
	if err != nil {
		t.Fatal(err)
	}

	if status.OS.Name == "" {
		t.Fatal("No OS")
	}
}

func TestNewSession(t *testing.T) {
	t.Parallel()
	if *runOnSauce {
		return
	}
	wd := &remoteWebDriver{capabilities: caps, executor: defaultExecutor}
	sid, err := wd.NewSession()
	defer wd.Quit()

	if err != nil {
		t.Fatalf("error in new session - %s", err)
	}

	if sid == "" {
		t.Fatal("Empty session id")
	}

	if wd.id != sid {
		t.Fatal("Session id mismatch")
	}
}

func TestCapabilities(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestCapabilities", t)
	defer wd.Quit()

	c, err := wd.Capabilities()
	if err != nil {
		t.Fatal(err)
	}

	if c["browserName"] != caps["browserName"] {
		t.Fatalf("bad browser name - %s", c["browserName"])
	}
}

func TestSetAsyncScriptTimeout(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestSetAsyncScriptTimeout", t)
	defer wd.Quit()

	err := wd.SetAsyncScriptTimeout(200)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSetImplicitWaitTimeout(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestSetImplicitWaitTimeout", t)
	defer wd.Quit()

	err := wd.SetImplicitWaitTimeout(200)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCurrentWindowHandle(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestCurrentWindowHandle", t)
	defer wd.Quit()

	handle, err := wd.CurrentWindowHandle()

	if err != nil {
		t.Fatal(err)
	}

	if handle == "" {
		t.Fatal("Empty handle")
	}
}

func TestWindowHandles(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestWindowHandles", t)
	defer wd.Quit()

	handles, err := wd.CurrentWindowHandle()
	if err != nil {
		t.Fatal(err)
	}

	if handles == "" {
		t.Fatal("No handles")
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestGet", t)
	defer wd.Quit()

	err := wd.Get(serverURL)
	if err != nil {
		t.Fatal(err)
	}

	newURL, err := wd.CurrentURL()
	if err != nil {
		t.Fatal(err)
	}

	if newURL != serverURL {
		t.Fatalf("%s != %s", newURL, serverURL)
	}
}

func TestNavigation(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestNavigation", t)
	defer wd.Quit()

	url1 := serverURL
	err := wd.Get(url1)
	if err != nil {
		t.Fatal(err)
	}

	url2 := serverURL + "other"
	err = wd.Get(url2)
	if err != nil {
		t.Fatal(err)
	}

	err = wd.Back()
	if err != nil {
		t.Fatal(err)
	}
	url, _ := wd.CurrentURL()
	if url != url1 {
		t.Fatalf("back got me to %s (expected %s)", url, url1)
	}
	err = wd.Forward()
	if err != nil {
		t.Fatal(err)
	}
	url, _ = wd.CurrentURL()
	if url != url2 {
		t.Fatalf("forward got me to %s (expected %s)", url, url2)
	}

	err = wd.Refresh()
	if err != nil {
		t.Fatal(err)
	}
	url, _ = wd.CurrentURL()
	if url != url2 {
		t.Fatalf("refresh got me to %s (expected %s)", url, url2)
	}
}

func TestTitle(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestTitle", t)
	defer wd.Quit()

	wd.Get(serverURL)

	title, err := wd.Title()
	if err != nil {
		t.Fatal(err)
	}

	expectedTitle := "Go Selenium Test Suite"
	if title != expectedTitle {
		t.Fatal("Bad title %s, should be %s", title, expectedTitle)
	}
}

func TestPageSource(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestPageSource", t)
	defer wd.Quit()

	err := wd.Get(serverURL)
	if err != nil {
		t.Fatal(err)
	}

	source, err := wd.PageSource()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(source, "The home page.") {
		t.Fatalf("Bad source\n%s", source)
	}
}

func TestFindElement(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestFindElement", t)
	defer wd.Quit()

	wd.Get(serverURL)
	elem, err := wd.FindElement(ByName, "q")
	if err != nil {
		t.Fatal(err)
	}

	we, ok := elem.(*remoteWE)
	if !ok {
		t.Fatal("Can't convert to *remoteWE")
	}

	if we.id == "" {
		t.Fatal("Empty element")
	}

	if we.parent != wd {
		t.Fatal("Bad parent")
	}
}

func TestFindElements(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestFindElements", t)
	defer wd.Quit()

	wd.Get(serverURL)
	elems, err := wd.FindElements(ByName, "q")
	if err != nil {
		t.Fatal(err)
	}

	if len(elems) != 1 {
		t.Fatal("Wrong number of elements %d (should be 1)", len(elems))
	}

	we, ok := elems[0].(*remoteWE)
	if !ok {
		t.Fatal("Can't convert to *remoteWE")
	}

	if we.id == "" {
		t.Fatal("Empty element")
	}

	if we.parent != wd {
		t.Fatal("Bad parent")
	}
}

func TestSendKeys(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestSendKeys", t)
	defer wd.Quit()

	wd.Get(serverURL)
	input, err := wd.FindElement(ByName, "q")
	if err != nil {
		t.Fatal(err)
	}
	err = input.SendKeys("golang\n")
	if err != nil {
		t.Fatal(err)
	}

	source, err := wd.PageSource()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(source, "The Go Programming Language") {
		t.Fatal("Can't find Go")
	}

	if !strings.Contains(source, "golang") {
		t.Fatal("Can't find search query in source")
	}

}

func TestClick(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestClick", t)
	defer wd.Quit()

	wd.Get(serverURL)
	input, err := wd.FindElement(ByName, "q")
	if err != nil {
		t.Fatal(err)
	}
	err = input.SendKeys("golang")
	if err != nil {
		t.Fatal(err)
	}

	button, err := wd.FindElement(ById, "submit")
	if err != nil {
		t.Fatal(err)
	}
	err = button.Click()
	if err != nil {
		t.Fatal(err)
	}

	source, err := wd.PageSource()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(source, "The Go Programming Language") {
		t.Fatal("Can't find Go")
	}
}

func TestGetCookies(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestGetCookies", t)
	defer wd.Quit()

	wd.Get(serverURL)
	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err)
	}

	if len(cookies) == 0 {
		t.Fatal("No cookies")
	}

	if cookies[0].Name == "" {
		t.Fatal("Empty cookie")
	}
}

func TestAddCookie(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestAddCookie", t)
	defer wd.Quit()

	wd.Get(serverURL)
	cookie := &Cookie{Name: "the nameless cookie", Value: "I have nothing"}
	err := wd.AddCookie(cookie)
	if err != nil {
		t.Fatal(err)
	}

	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cookies {
		if (c.Name == cookie.Name) && (c.Value == cookie.Value) {
			return
		}
	}

	t.Fatal("Can't find new cookie")
}

func TestDeleteCookie(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestDeleteCookie", t)
	defer wd.Quit()

	wd.Get(serverURL)
	cookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err)
	}
	if len(cookies) == 0 {
		t.Fatal("No cookies")
	}
	err = wd.DeleteCookie(cookies[0].Name)
	if err != nil {
		t.Fatal(err)
	}
	newCookies, err := wd.GetCookies()
	if err != nil {
		t.Fatal(err)
	}
	if len(newCookies) != len(cookies)-1 {
		t.Fatal("Cookie not deleted")
	}

	for _, c := range newCookies {
		if c.Name == cookies[0].Name {
			t.Fatal("Deleted cookie found")
		}
	}
}

func TestLocation(t *testing.T) {
	wd := newRemote("TestLocation", t)
	defer wd.Quit()

	wd.Get(serverURL)
	button, err := wd.FindElement(ById, "submit")
	if err != nil {
		t.Fatal(err)
	}

	loc, err := button.Location()
	if err != nil {
		t.Fatal(err)
	}

	if (loc.X == 0) || (loc.Y == 0) {
		t.Fatalf("Bad location: %v\n", loc)
	}
}

func TestLocationInView(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestLocationInView", t)
	defer wd.Quit()

	wd.Get(serverURL)
	button, err := wd.FindElement(ById, "submit")
	if err != nil {
		t.Fatal(err)
	}

	loc, err := button.LocationInView()
	if err != nil {
		t.Fatal(err)
	}

	if (loc.X == 0) || (loc.Y == 0) {
		t.Fatalf("Bad location: %v\n", loc)
	}
}

func TestSize(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestSize", t)
	defer wd.Quit()

	wd.Get(serverURL)
	button, err := wd.FindElement(ById, "submit")
	if err != nil {
		t.Fatal(err)
	}

	size, err := button.Size()
	if err != nil {
		t.Fatal(err)
	}

	if (size.Width == 0) || (size.Height == 0) {
		t.Fatalf("Bad size: %v\n", size)
	}
}

func TestExecuteScript(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestExecuteScript", t)
	defer wd.Quit()

	script := "return arguments[0] + arguments[1]"
	args := []interface{}{1, 2}
	reply, err := wd.ExecuteScript(script, args)
	if err != nil {
		t.Fatal(err)
	}

	result, ok := reply.(float64)
	if !ok {
		t.Fatal("Not an int reply")
	}

	if result != 3 {
		t.Fatal("Bad result %d (expected 3)", result)
	}
}

func TestScreenshot(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestScreenshot", t)
	defer wd.Quit()

	wd.Get(serverURL)
	data, err := wd.Screenshot()
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Fatal("Empty reply")
	}
}

func TestIsSelected(t *testing.T) {
	t.Parallel()
	wd := newRemote("TestIsSelected", t)
	defer wd.Quit()

	wd.Get(serverURL)
	elem, err := wd.FindElement(ById, "chuk")
	if err != nil {
		t.Fatal("Can't find element")
	}
	selected, err := elem.IsSelected()
	if err != nil {
		t.Fatal("Can't get selection")
	}

	if selected {
		t.Fatal("Already selected")
	}

	err = elem.Click()
	if err != nil {
		t.Fatal("Can't click")
	}

	selected, err = elem.IsSelected()
	if err != nil {
		t.Fatal("Can't get selection")
	}

	if !selected {
		t.Fatal("Not selected")
	}
}

// Test server

var homePage = `
<html>
<head>
	<title>Go Selenium Test Suite</title>
</head>
<body>
	The home page. <br />
	<form action="/search">
		<input name="q" /> <input type="submit" id="submit"/> <br />
		<input id="chuk" type="checkbox" /> A checkbox.
	</form>
</body>
</html>
`

var otherPage = `
<html>
<head>
	<title>Go Selenium Test Suite - Other Page</title>
</head>
<body>
	The other page.
</body>
</html>
`

var searchPage = `
<html>
<head>
	<title>Go Selenium Test Suite - Search Page</title>
</head>
<body>
	You searched for "%s". I'll pretend I've found:
	<p>
	"The Go Programming Language"
	</p>
</body>
</html>
`

var pages = map[string]string{
	"/":       homePage,
	"/other":  otherPage,
	"/search": searchPage,
}

func handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	page, ok := pages[path]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if path == "/search" {
		r.ParseForm()
		page = fmt.Sprintf(page, r.Form["q"][0])
	}
	// Some cookies for the tests
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("cookie-%d", i)
		value := fmt.Sprintf("value-%d", i)
		http.SetCookie(w, &http.Cookie{Name: name, Value: value})
	}
	fmt.Fprintf(w, page)
}

var serverPort = ":4793"
var serverURL = "http://localhost" + serverPort + "/"

func init() {
	go func() {
		http.HandleFunc("/", handler)
		http.ListenAndServe(serverPort, nil)
	}()
}
