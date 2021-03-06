package conformance

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

const (
	htmlTemplate string = `<html>
  <head>
    <title>OCI Distribution Conformance Tests</title>
    <style>
	  body {
	    padding: 30px;
	  }
      .result {
        padding: 1.25em 0 .25em 2em;
        border: 2px solid #e2e2e2;
        border-radius: 5px;
        margin-top: 10px;
      }
      .red {
        background: #ffc8c8;
      }
      pre.fail-message {
        background: #f9a5a5;
        padding: 20px;
        display: inline-block;
        border-radius: 4px;
        font-size: 1.25em;
      }
      .green {
        background: #c8ffc8;
        padding: 1.25em 0 1.25em 2em;
      }
      .toggle {
        border: 2px solid #3e3e3e;
        cursor: pointer;
        width: 1em;
        text-align: center;
        font-weight: bold;
        display: inline;
        font-family: monospace;
        padding: 0 .25em 0 .25em;
        margin: 1em 1em 1em 0;
        font-size: 12pt;
        color: #3e3e3e;
        border-radius: 3px;
      }
      pre.pre-box {
        background: white;
        padding: 10px;
        border: 1px solid gray;
        display: inline-block;
        border-radius: 4px;
        width: 97%;
        font-size: 1.25em;
        overflow-x: auto;
        max-height: 60em;
        overflow-y: auto;
      }
	  .summary {
		width: 75%;
		height: auto;
		padding: 0 0 .5em 0;
		border: 1px solid grey;
	  }
	  .summary-bullet {
		width: 100%;
		height: auto;
		display: flex;
		padding: .5em .1em .1em .5em;
	  }
	  .bullet-left {
		width: 20%;
		font-weight: bold;
		font-size: 1.1em;
	  }
	  .bullet-right {
		width: auto;
	  }
	  .quick-summary {
		width: 70%;
		display: flex;
		margin: 0 auto 0 0;
		font-weight: bold;
		font-size: 1.2em;
	  }
	  .darkgreen {
		color: green;
		padding: 0 2em 0 0;
	  }
	  .darkred {
		color: red;
	  }
	  .meter {
		border: 1px solid black;
		margin: 0 .5em 0 auto;
		display: flex;
		height: 25px;
		width: 50%;
	  }
	  .meter-green {
		height: 100%;
		background: green;
		width: {{ .PercentPassed -}}%;
	  }
	  .meter-red {
		height: 100%;
		background: red;
		width: {{ .PercentFailed -}}%;
	  }
    </style>
    <script>
      function toggleOutput(id) {
		var elem = document.getElementById(id);
		var button = document.getElementById(id + "-button");
		if (elem.style['display'] === 'block') {
			button.innerHTML = "+";
			elem.style['display'] = 'none';
		} else {
			button.innerHTML = "-";
			elem.style['display'] = 'block';
		}
	}
    </script>
  </head>
  <body>
    <h1>OCI Distribution Conformance Tests</h1>
<div class="summary">
	<div class="summary-bullet">
		<div class="bullet-left">Summary:</div>
		<div class="quick-summary">
			{{- if not .AllFailed -}}
				<span class="darkgreen">
				{{- if .AllPassed -}}All {{ end -}}{{ .SuiteSummary.NumberOfPassedSpecs }} passed</span>
			{{- end -}}
			{{- if not .AllPassed -}}
				<span class="darkred">
				{{- if .AllFailed -}}All {{ end -}}{{ .SuiteSummary.NumberOfFailedSpecs }} failed</span>
			{{- end -}}
		  <div class="meter">
			<div class="meter-green"></div><div class="meter-red"></div>
		  </div>
		</div>
	</div>
	<div class="summary-bullet">
		<div class="bullet-left">Start time:</div>
		<div class="bullet-right">{{ .StartTimeString }}</div>
	</div>
	<div class="summary-bullet">
		<div class="bullet-left">End time:</div>
		<div class="bullet-right">{{ .EndTimeString }}</div>
	</div>
	<div class="summary-bullet">
		<div class="bullet-left">Elapsed time:</div>
		<div class="bullet-right">{{ .RunTime }}</div>
	</div>
	<div class="summary-bullet">
		<div class="bullet-left">Test version:</div>
		<div class="bullet-right">{{ .Version }}</div>
	</div>
	<div class="summary-bullet">
		<div class="bullet-left">Configuration:</div>
		<div class="bullet-right">
			{{ range $i, $s := .EnvironmentVariables }}
				{{ $s }}<br />
			{{ end }}
		</div>
	</div>
</div>
    <div>
      {{with .SpecSummaryMap}}
        {{$x := .M}}
        {{range $i, $k := .Keys}}
          <h2>{{$k}}</h2>
          {{$v := index $x $k}}
          {{range $z, $s := $v}}
            {{if eq $s.State 4}}
              <div class="result red">
                <div id="output-box-{{$s.ID}}-button" class="toggle"
                  onclick="javascript:toggleOutput('output-box-{{$s.ID}}')">+</div>
                <h3 style="display: inline;">{{$s.Title}}</h3>
                <br>
                <div>
                  <div id="output-box-{{$s.ID}}" style="display: none;">
                    <pre class="pre-box">{{$s.CapturedOutput}}</pre>
                  </div>
                </div>
                <pre class="fail-message">{{$s.Failure.Message}}</pre>
                <br>
              </div>
    </div>
            {{else if eq $s.State 3}}
              <div class="result green">
                <div id="output-box-{{$s.ID}}-button" class="toggle"
                  onclick="javascript:toggleOutput('output-box-{{$s.ID}}')">+</div>
                <h3 style="display: inline;">{{$s.Title}}</h3>
                <br>
                <div id="output-box-{{$s.ID}}" style="display: none;">
                  <pre class="pre-box">{{$s.CapturedOutput}}</pre>
                </div>
			  </div>
            {{end}}
          {{end}}
        {{end}}
      {{end}}
  </body>
</html>
`
)

type (
	summaryMap struct {
		M    map[string][]specSnapshot
		Keys []string
		Size int
	}

	specSnapshot struct {
		types.SpecSummary
		ID    int
		Title string
	}

	httpDebugWriter struct {
		CapturedOutput []string
		debug          bool
	}

	httpDebugLogger struct {
		l *log.Logger
		w io.Writer
	}
)

func (sm *summaryMap) Add(key string, sum *specSnapshot) {
	sm.M[key] = append(sm.M[key], *sum)
	sm.Size++

	if !sm.containsKey(key) {
		sm.Keys = append(sm.Keys, key)
	}
}

func (sm *summaryMap) containsKey(key string) bool {
	var containsKey bool
	for _, k := range sm.Keys {
		if k == key {
			containsKey = true
			break
		}
	}
	return containsKey
}

func newSpecSnapshot(sum *types.SpecSummary, id int) *specSnapshot {
	return &specSnapshot{SpecSummary: *sum, Title: sum.ComponentTexts[3], ID: id}
}

func newHTTPDebugWriter(debug bool) *httpDebugWriter {
	return &httpDebugWriter{debug: debug}
}

func (writer *httpDebugWriter) Write(b []byte) (int, error) {
	s := string(b)
	writer.CapturedOutput = append(writer.CapturedOutput, s)
	if writer.debug {
		fmt.Println(s)
	}

	return len(b), nil
}

func newHTTPDebugLogger(f io.Writer) *httpDebugLogger {
	debugLogger := &httpDebugLogger{w: f, l: log.New(f, "", log.Ldate|log.Lmicroseconds)}
	return debugLogger
}

func (l *httpDebugLogger) Errorf(format string, v ...interface{}) {
	l.output("ERROR "+format, v...)
}

func (l *httpDebugLogger) Warnf(format string, v ...interface{}) {
	l.output("WARN "+format, v...)
}

func (l *httpDebugLogger) Debugf(format string, v ...interface{}) {
	re := regexp.MustCompile("(?i)(\"?\\w*(authorization|token|state)\\w*\"?(:|=)\\s*)(\")?\\s*((bearer|basic)? )?[^\\s&\"]*(\")?")
	format = re.ReplaceAllString(format, "$1$4$5*****$7")
	l.output("DEBUG "+format, v...)
}

func (l *httpDebugLogger) output(format string, v ...interface{}) {
	if len(v) == 0 {
		l.l.Print(format)
		return
	}
	_, err := l.w.Write([]byte(fmt.Sprintf(format, v...)))
	if err != nil {
		l.Errorf(err.Error())
	}
}

type (
	HTMLReporter struct {
		htmlReportFilename   string
		SpecSummaryMap       summaryMap
		EnvironmentVariables []string
		SuiteSummary         *types.SuiteSummary
		debugLogger          *httpDebugWriter
		debugIndex           int
		PercentPassed        int
		PercentFailed        int
		StartTime            time.Time
		StartTimeString      string
		EndTime              time.Time
		EndTimeString        string
		RunTime              string
		AllPassed            bool
		AllFailed            bool
		Version              string
	}
)

func newHTMLReporter(htmlReportFilename string) *HTMLReporter {
	return &HTMLReporter{
		htmlReportFilename: htmlReportFilename,
		debugLogger:        httpWriter,
		SpecSummaryMap:     summaryMap{M: make(map[string][]specSnapshot)},
	}
}

func (reporter *HTMLReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	b := new(bytes.Buffer)
	for _, co := range httpWriter.CapturedOutput[reporter.debugIndex:] {
		b.WriteString(co)
		b.WriteString("\n")
	}
	specSummary.CapturedOutput = b.String()

	header := specSummary.ComponentTexts[2]
	summary := newSpecSnapshot(specSummary, reporter.SpecSummaryMap.Size)
	reporter.SpecSummaryMap.Add(header, summary)
	reporter.debugIndex = len(reporter.debugLogger.CapturedOutput)
}

func (reporter *HTMLReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	reporter.EndTime = time.Now()
	reporter.EndTimeString = reporter.EndTime.Format("Jan 2 15:04:05.000 -0700 MST")
	reporter.RunTime = reporter.EndTime.Sub(reporter.StartTime).String()
	reporter.PercentPassed = int(float64(summary.NumberOfPassedSpecs) / float64(summary.NumberOfTotalSpecs) * 100)
	reporter.PercentFailed = 100 - reporter.PercentPassed
	reporter.SuiteSummary = summary
	reporter.AllPassed = summary.NumberOfPassedSpecs == (summary.NumberOfTotalSpecs - summary.NumberOfSkippedSpecs)
	reporter.AllFailed = summary.NumberOfFailedSpecs == (summary.NumberOfTotalSpecs - summary.NumberOfSkippedSpecs)

	t, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		log.Fatal(err)
	}

	htmlReportFilenameAbsPath, err := filepath.Abs(reporter.htmlReportFilename)
	if err != nil {
		log.Fatal(err)
	}

	htmlReportFile, err := os.Create(htmlReportFilenameAbsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer htmlReportFile.Close()

	err = t.ExecuteTemplate(htmlReportFile, "report", &reporter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("HTML report was created: %s", htmlReportFilenameAbsPath)
}

//unused by HTML reporter
func (reporter *HTMLReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	varsToCheck := []string{
		"OCI_ROOT_URL",
		"OCI_NAMESPACE",
		"OCI_DEBUG",
		"OCI_PASSWORD",
		"OCI_USERNAME",
	}
	for _, v := range varsToCheck {
		var replacement string
		if envVar := os.Getenv(v); envVar != "" {
			replacement = envVar
			if strings.Contains(v, "PASSWORD") || strings.Contains(v, "USERNAME") {
				replacement = "*****"
			}
		} else {
			continue
		}
		reporter.EnvironmentVariables = append(reporter.EnvironmentVariables,
			fmt.Sprintf("%s=%s", v, replacement))
	}

	reporter.StartTime = time.Now()
	reporter.StartTimeString = reporter.StartTime.Format("Jan 2 15:04:05.000 -0700 MST")

	reporter.Version = Version
}

func (reporter *HTMLReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (reporter *HTMLReporter) SpecWillRun(specSummary *types.SpecSummary) {
}

func (reporter *HTMLReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
}
