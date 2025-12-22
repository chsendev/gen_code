<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="{{.MapperPackage}}.{{.ClassName}}Mapper">

    <!-- 通用查询映射结果 -->
    <resultMap id="BaseResultMap" type="{{.EntityPackage}}.{{.ClassName}}">
        {{range .Table.Fields}}
        <result column="{{.ColumnName}}" property="{{.FieldName}}" />
        {{end}}
    </resultMap>

    <!-- 通用查询结果列 -->
    <sql id="Base_Column_List">
        {{range $index, $field := .Table.Fields}}
        {{.ColumnName}}{{if ne $index (sub (len $.Table.Fields) 1)}},{{end}}
        {{end}}
    </sql>

</mapper>
