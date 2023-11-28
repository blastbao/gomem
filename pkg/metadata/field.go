// Copyright 2019 Nick Poorman
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metadata

import "github.com/apache/arrow/go/arrow"

const (
	originalTypeKey = "GOMEM_DATAFRAME_ORIGINAL_TYPE"
	mapConstant     = "MAP"
	logicalTypeKey  = "LogicalType"
)

func AppendOriginalTypeMetadata(metadata arrow.Metadata, value string) arrow.Metadata {
	keys := append(metadata.Keys(), originalTypeKey, logicalTypeKey)
	values := append(metadata.Values(), value, value)
	return arrow.NewMetadata(keys, values)
}

func AppendOriginalMapTypeMetadata(metadata arrow.Metadata) arrow.Metadata {
	return AppendOriginalTypeMetadata(metadata, mapConstant)
}

// OriginalMapTypeMetadataExists 判断当前 field 是否是 map 类型
func OriginalMapTypeMetadataExists(metadata arrow.Metadata) bool {
	// 从 metadata 中获取 'LogicalType' 的值是否为 'MAP'
	if value, ok := metadataValue(metadata, logicalTypeKey); ok {
		return value == mapConstant
	}
	// 从 metadata 中获取 'GOMEM_DATAFRAME_ORIGINAL_TYPE' 的值是否为 'MAP'
	if value, ok := metadataValue(metadata, originalTypeKey); ok {
		return value == mapConstant
	}
	return false
}

// 从 metadata 中获取 key 对应的 value ，不存在返回 _, false
func metadataValue(metadata arrow.Metadata, key string) (string, bool) {
	// 查找 key 的 idx
	idx := metadata.FindKey(key)
	if idx == -1 {
		return "", false
	}
	// 根据 idx 取出 value
	return metadata.Values()[idx], true
}
