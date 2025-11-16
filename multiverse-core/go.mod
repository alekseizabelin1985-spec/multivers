module multiverse-core

go 1.25

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/minio/minio-go/v7 v7.0.95
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
	github.com/segmentio/kafka-go v0.4.49
	github.com/xeipuuv/gojsonschema v1.2.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
)

require (
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/gorilla/websocket v1.5.3
	github.com/minio/crc64nvme v1.1.1 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/stretchr/testify v1.10.0
	github.com/tinylib/msgp v1.5.0 // indirect
)

require (
	//github.com/amikos-tech/chroma-go v0.1.4
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
)

// Требуется для chroma-go v0.2.5
replace github.com/milvus-io/milvus-sdk-go/v2 => github.com/milvus-io/milvus-sdk-go/v2 v2.4.4

// Критически важные замены для Windows/CGO
// Отключаем проблемные зависимости
//replace github.com/yalue/onnxruntime_go => github.com/yalue/onnxruntime_go v0.0.0-20230101000000-000000000000
//replace github.com/amikos-tech/chroma-go/pkg/tokenizers/libtokenizerscompiler => github.com/amikos-tech/chroma-go/pkg/tokenizers/libtokenizerscompiler v0.0.0-20230101000000-000000000000
// Заменяем проблемные зависимости на фиктивные
//replace github.com/yalue/onnxruntime_go => ./fake_deps
//replace github.com/amikos-tech/chroma-go/pkg/tokenizers/libtokenizerscompiler => ./fake_deps
//replace github.com/amikos-tech/chroma-go/pkg/embeddings/default_ef => ./fake_deps
//exclude github.com/amikos-tech/chroma-go/pkg/tokenizers/libtokenizerscompiler v0.2.5
