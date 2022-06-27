package app

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_splitPath(t *testing.T) {
	Convey("detects errors", t, func() {
		Convey("rejects absolute paths", func() {
			_, _, err := splitPath("/a/path")
			So(err, ShouldBeError, "path must not be absolute")
		})
		Convey("rejects missing filename", func() {
			_, _, err := splitPath("dir/")
			So(err, ShouldBeError, "path must include filename")
		})
	})

	Convey("returns ok", t, func() {
		var tests = map[string]struct {
			path string
			dir  string
			name string
		}{
			"just filename": {
				path: "filename",
				dir:  "",
				name: "filename",
			},
			"dir and filename": {
				path: "dir/file",
				dir:  "dir",
				name: "file",
			},
			"long dir and filename": {
				path: "multi/segment/dir/filename",
				dir:  "multi/segment/dir",
				name: "filename",
			},
		}
		for name, test := range tests {
			Convey(name, func() {
				dir, name, err := splitPath(test.path)
				So(err, ShouldBeNil)
				So(dir, ShouldEqual, test.dir)
				So(name, ShouldEqual, test.name)
			})
		}
	})

}

func Test_isRegular(t *testing.T) {
	Convey("detects types of files", t, func() {
		var tests = map[string]struct {
			name string
			want bool
		}{
			"directory":    {".", false},
			"regular file": {"app_test.go", true},
			"device file":  {"/dev/null", false},
		}

		for desc, test := range tests {
			Convey(desc, func() {
				info, err := os.Stat(test.name)
				So(err, ShouldBeNil)
				got := isRegular(info)
				So(got, ShouldEqual, test.want)
			})
		}
	})
}
