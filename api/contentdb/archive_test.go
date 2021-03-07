package contentdb

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"testing"
)

// helper variables to test cases
var (
	bytesMesecons []byte      = mustReadFile("testdata/mesecons.zip")
	zipMesecons   *zip.Reader = mustZip(bytesMesecons)

	bytes3darmor []byte      = mustReadFile("testdata/3d_armor.zip")
	zip3darmor   *zip.Reader = mustZip(bytes3darmor)

	bytesSfinv []byte      = mustReadFile("testdata/sfinv.zip")
	zipSfinv   *zip.Reader = mustZip(bytesSfinv)

	bytesTelegram []byte      = mustReadFile("testdata/telegram.zip")
	zipTelegram   *zip.Reader = mustZip(bytesTelegram)

	bytesRespawn []byte      = mustReadFile("testdata/respawn.zip")
	zipRespawn   *zip.Reader = mustZip(bytesRespawn)

	bytesNestedMod []byte      = mustReadFile("testdata/nestedmod.zip")
	zipNestedMod   *zip.Reader = mustZip(bytesNestedMod)

	bytesInvalidModpack []byte      = mustReadFile("testdata/invalidmodpack.zip")
	zipInvalidModpack   *zip.Reader = mustZip(bytesInvalidModpack)
)

func mustReadFile(f string) []byte {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	return b
}

func mustZip(b []byte) *zip.Reader {
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		panic(err)
	}
	return z
}

func TestPackageArchive_FindFile(t *testing.T) {
	type fields struct {
		b        *bytes.Buffer
		z        *zip.Reader
		contents []string
	}
	type args struct {
		name string
		max  int
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantCount int
		wantDir   string
	}{
		{
			name:      "should find init.lua inside a valid mod",
			fields:    fields{z: zipSfinv},
			args:      args{"init.lua", 1},
			wantCount: 1,
			wantDir:   "sfinv/",
		},
		{
			name:      "should find all init.lua files inside a modpack",
			fields:    fields{z: zip3darmor},
			args:      args{"init.lua", 0},
			wantCount: 7,
			wantDir:   "3d_armor/wieldview/",
		},
		{
			name:      "should find only first init.lua when max is 1",
			fields:    fields{z: zip3darmor},
			args:      args{"init.lua", 1},
			wantCount: 1,
			wantDir:   "3d_armor/3d_armor/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PackageArchive{
				b:        tt.fields.b,
				z:        tt.fields.z,
				contents: tt.fields.contents,
			}
			gotCount, gotDir := p.FindFile(tt.args.name, tt.args.max)
			if gotCount != tt.wantCount {
				t.Errorf("PackageArchive.FindFile() gotCount = %v, want %v", gotCount, tt.wantCount)
			}
			if gotDir != tt.wantDir {
				t.Errorf("PackageArchive.FindFile() gotDir = %v, want %v", gotDir, tt.wantDir)
			}
		})
	}
}

func TestPackageArchive_Type(t *testing.T) {
	type fields struct {
		b        *bytes.Buffer
		z        *zip.Reader
		contents []string
	}
	tests := []struct {
		name   string
		fields fields
		want   ArchiveType
	}{
		{
			name:   "should detect valid mod as MOD",
			fields: fields{b: bytes.NewBuffer(bytesSfinv), z: zipSfinv},
			want:   Mod,
		},
		{
			name:   "should detect valid modpack as MODPACK",
			fields: fields{b: bytes.NewBuffer(bytesMesecons), z: zipMesecons},
			want:   Modpack,
		},
		{
			name:   "should detect mod without mod.conf",
			fields: fields{b: bytes.NewBuffer(bytesTelegram), z: zipTelegram},
			want:   Mod,
		},
		{
			name:   "should detect mod with nested dirs before init.lua",
			fields: fields{b: bytes.NewBuffer(bytesNestedMod), z: zipNestedMod},
			want:   Mod,
		},
		{
			name:   "should detect invalid modpack with multiple mods inside",
			fields: fields{b: bytes.NewBuffer(bytesInvalidModpack), z: zipInvalidModpack},
			want:   Invalid,
		},
		{
			name:   "should detect valid mod at the root of zip folder",
			fields: fields{b: bytes.NewBuffer(bytesRespawn), z: zipRespawn},
			want:   Mod,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PackageArchive{
				b:        tt.fields.b,
				z:        tt.fields.z,
				contents: tt.fields.contents,
			}
			if got := p.Type(); got != tt.want {
				t.Errorf("PackageArchive.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
