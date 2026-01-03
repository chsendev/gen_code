@@Meta.Output="/src/main/java/{{.EntityPackage | replace "." "/"}}/{{.ClassName}}.java"

package {{.EntityPackage}};

{{if .EnableLombok}}import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.experimental.Accessors;
{{end}}
import com.baomidou.mybatisplus.annotation.TableName;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableField;
import java.io.Serializable;
import java.math.BigDecimal;
import java.util.Date;

/**
 * {{.Table.TableComment}}
 * @author {{.Author}}
 * @date {{.Date}}
 */
{{if .EnableLombok}}@Data
@EqualsAndHashCode(callSuper = false)
@Accessors(chain = true)
{{end}}
@TableName("{{.Table.TableName}}")
public class {{.ClassName}} implements Serializable {

    private static final long serialVersionUID = 1L;

    {{range .Table.Fields}}
    {{if .IsPrimaryKey}}@TableId("{{.ColumnName}}")
    {{else}}@TableField("{{.ColumnName}}")
    {{end}}
    private {{.JavaType}} {{.FieldName}};

    {{end}}
}
