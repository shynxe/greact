package build

import (
	"reflect"
	"testing"

	"github.com/shynxe/greact/config"
)

func Test_getJSFiles(t *testing.T) {
	type args struct {
		filenames []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test getJSFiles",
			args: args{
				filenames: []string{
					"index.js",
					"index.css",
					"index.html",
				},
			},
			want: []string{
				"index.js",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getJSFiles(tt.args.filenames); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getJSFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getWebpackConfig(t *testing.T) {
	getJSSourceFiles = func() []string {
		return []string{
			"index.js",
			"about.js",
		}
	}

	type args struct {
		userConfig config.Config
	}
	tests := []struct {
		name string
		args args
		want WebpackConfig
	}{
		{
			name: "Test getWebpackConfig",
			args: args{
				userConfig: config.Config{
					ClientPath:   "client",
					SourceFolder: "src",
					BuildFolder:  "build",
					StaticFolder: "static",
					PublicPath:   "/",
				},
			},
			want: WebpackConfig{
				EntryPoints:        "index: path.join(__dirname, 'src', 'index.js'),\n\t\tabout: path.join(__dirname, 'src', 'about.js'),\n\t\t",
				HtmlWebpackPlugins: "new HtmlWebpackPlugin({\n\t\t\ttemplate: path.join(__dirname, 'build', '.greact-template.html'),\n\t\t\tfilename: 'index.html',\n\t\t\tchunks: ['hydrate', 'index'],\n\t\t\tpublicPath: '/',\n\t\t}),\n\t\tnew HtmlWebpackPlugin({\n\t\t\ttemplate: path.join(__dirname, 'build', '.greact-template.html'),\n\t\t\tfilename: 'about.html',\n\t\t\tchunks: ['hydrate', 'about'],\n\t\t\tpublicPath: '/',\n\t\t}),\n\t\t",
				BuildFolder:        "build",
				StaticFolder:       "static",
				PublicPath:         "/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getWebpackConfig(tt.args.userConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getWebpackConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
