@@Meta.Output="/src/main/java/{{.ServicePackage | replace "." "/"}}/I{{.ClassName}}Service.java"

package {{.ServicePackage}};

import com.baomidou.mybatisplus.extension.service.IService;
import {{.EntityPackage}}.{{.ClassName}};

/**
 * {{.Table.TableComment}}Service接口
 * @author {{.Author}}
 * @date {{.Date}}
 */
public interface I{{.ClassName}}Service extends IService<{{.ClassName}}> {

}
