package templating

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_TemplateDescription(t *testing.T) {
	Convey("Test templates", t, func() {
		var Name = "TestName"
		var Desc = "" +
			"Trigger name: {{.Trigger.Name}}\n" +
			"{{range $v := .Events }}\n" +
			"Metric: {{$v.Metric}}\n" +
			"MetricElements: {{$v.MetricElements}}\n" +
			"Timestamp: {{$v.Timestamp}}\n" +
			"Value: {{$v.Value}}\n" +
			"State: {{$v.State}}\n" +
			"{{end}}\n" +
			"https://grafana.yourhost.com/some-dashboard" +
			"{{ range $i, $v := .Events }}{{ if ne $i 0 }}&{{ else }}?" +
			"{{ end }}var-host={{ $v.Metric }}{{ end }}"

		var testUnixTime = time.Now().Unix()
		var events = []Event{{Metric: "1", Timestamp: testUnixTime}, {Metric: "2", Timestamp: testUnixTime}}

		Convey("Test nil data", func() {
			expected, err := Populate(Name, Desc, nil)
			if err != nil {
				println("Error:", err.Error())
			}
			So(err, ShouldBeNil)
			So(`Trigger name: TestName

https://grafana.yourhost.com/some-dashboard`,
				ShouldResemble, expected)
		})

		Convey("Test data", func() {
			expected, err := Populate(Name, Desc, events)
			So(err, ShouldBeNil)
			So(fmt.Sprintf("Trigger name: TestName\n\nMetric: 1\nMetricElements: []\nTimestamp: %d\nValue: &lt;nil&gt;"+
				"\nState: \n\nMetric: 2\nMetricElements: []\nTimestamp: %d\nValue: &lt;nil&gt;"+
				"\nState: \n\nhttps://grafana.yourhost.com/some-dashboard?var-host=1&var-host=2", testUnixTime, testUnixTime),
				ShouldResemble, expected)
		})

		Convey("Test description without templates", func() {
			anotherText := "Another text"
			Desc = anotherText

			expected, err := Populate(Name, Desc, events)
			So(err, ShouldBeNil)
			So(anotherText, ShouldEqual, expected)
		})

		Convey("Test method Date", func() {
			formatDate := time.Unix(testUnixTime, 0).Format(eventTimeFormat)
			actual := fmt.Sprintf("%s | %s |", formatDate, formatDate)
			Desc = "{{ range .Events }}{{ date .Timestamp }} | {{ end }}"

			expected, err := Populate(Name, Desc, events)
			So(err, ShouldBeNil)
			So(actual, ShouldEqual, expected)
		})

		Convey("Test method formatted Date", func() {
			formatedDate := time.Unix(testUnixTime, 0).Format("2006-01-02 15:04:05")
			actual := fmt.Sprintf("%s | %s |", formatedDate, formatedDate)
			Desc = "{{ range .Events }}{{ formatDate .Timestamp \"2006-01-02 15:04:05\" }} | {{ end }}"

			expected, err := Populate(Name, Desc, events)
			So(err, ShouldBeNil)
			So(actual, ShouldEqual, expected)
		})

		Convey("Test method decrease and increase Date", func() {
			var timeOffset int64 = 300

			Convey("Date increase", func() {
				increase := testUnixTime + timeOffset
				actual := fmt.Sprintf("%d | %d |", increase, increase)
				Desc = fmt.Sprintf("{{ range .Events }}{{ .TimestampIncrease %d }} | {{ end }}", timeOffset)

				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, expected)
			})

			Convey("Date decrease", func() {
				increase := testUnixTime - timeOffset
				actual := fmt.Sprintf("%d | %d |", increase, increase)
				Desc = fmt.Sprintf("{{ range .Events }}{{ .TimestampDecrease %d }} | {{ end }}", timeOffset)

				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So(actual, ShouldEqual, expected)
			})
		})

		Convey("Bad functions", func() {
			var timeOffset int64 = 300

			Convey("Non-existent function", func() {
				Desc = fmt.Sprintf("{{ range .Events }}{{ decrease %d }} | {{ end }}", timeOffset)

				expected, err := Populate(Name, Desc, events)
				So(err, ShouldNotBeNil)
				So(Desc, ShouldEqual, expected)
			})

			Convey("Non-existent method", func() {
				Desc = fmt.Sprintf("{{ range .Events }}{{ .Decrease %d }} | {{ end }}", timeOffset)

				expected, err := Populate(Name, Desc, events)
				So(err, ShouldNotBeNil)
				So(Desc, ShouldEqual, expected)
			})

			Convey("Bad parameters", func() {
				Desc = "{{ date \"bad\" }} "

				expected, err := Populate(Name, Desc, events)
				So(err, ShouldNotBeNil)
				So(Desc, ShouldEqual, expected)
			})

			Convey("No parameters", func() {
				Desc = "{{ date }} "

				expected, err := Populate(Name, Desc, events)
				So(err, ShouldNotBeNil)
				So(Desc, ShouldEqual, expected)
			})
		})

		Convey("Test strings functions", func() {
			Convey("Test replace", func() {
				Desc = "{{ stringsReplace \"my.metrics.path\" \".\" \"_\" -1 }} "
				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So("my_metrics_path", ShouldEqual, expected)
			})

			Convey("Test replace limited to 1", func() {
				Desc = "{{ stringsReplace \"my.metrics.path\" \".\" \"_\" 1 }} "
				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So("my_metrics.path", ShouldEqual, expected)
			})

			Convey("Test trim suffix", func() {
				Desc = "{{ stringsTrimSuffix \"my.metrics.path\" \".path\" }} "
				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So("my.metrics", ShouldEqual, expected)
			})

			Convey("Test trim prefix", func() {
				Desc = "{{ stringsTrimPrefix \"my.metrics.path\" \"my.\" }} "
				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So("metrics.path", ShouldEqual, expected)
			})

			Convey("Test lower case", func() {
				Desc = "{{ stringsToLower \"MY.PATH\" }} "
				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So("my.path", ShouldEqual, expected)
			})

			Convey("Test upper case", func() {
				Desc = "{{ stringsToUpper \"my.path\" }} "
				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So("MY.PATH", ShouldEqual, expected)
			})
		})

		Convey("Test some sprig functions", func() {
			Convey("Test upper", func() {
				Desc = "{{ \"hello!\" | upper}} "
				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So("HELLO!", ShouldEqual, expected)
			})

			Convey("Test upper repeat", func() {
				Desc = "{{ \"hello!\" | upper | repeat 5 }} "
				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So("HELLO!HELLO!HELLO!HELLO!HELLO!", ShouldEqual, expected)
			})

			Convey("Test list uniq without", func() {
				Desc = "{{ without (list 1 3 3 2 2 2 4 4 4 4 1 | uniq) 4 }} "
				expected, err := Populate(Name, Desc, events)
				So(err, ShouldBeNil)
				So("[1 3 2]", ShouldEqual, expected)
			})
		})
	})
}
