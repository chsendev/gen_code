@@Meta.Output="/src/main/java/{{.MapperPackage | replace "." "/"}}/{{.ClassName}}Mapper.java"

package {{.MapperPackage}};

import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import {{.EntityPackage}}.{{.ClassName}};

/**
 * {{.Table.TableComment}}Mapper接口
 * @author {{.Author}}
 * @date {{.Date}}
 */
public interface {{.ClassName}}Mapper extends BaseMapper<{{.ClassName}}> {

}
