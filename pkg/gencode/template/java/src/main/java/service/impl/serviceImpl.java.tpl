@@Meta.Output="/src/main/java/{{.ServicePackage | replace "." "/"}}/impl/{{.ClassName}}ServiceImpl.java"

package {{.ServicePackage}}.impl;

import org.springframework.stereotype.Service;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import {{.EntityPackage}}.{{.ClassName}};
import {{.MapperPackage}}.{{.ClassName}}Mapper;
import {{.ServicePackage}}.I{{.ClassName}}Service;

/**
 * {{.Table.TableComment}}Service实现类
 * @author {{.Author}}
 * @date {{.Date}}
 */
@Service
public class {{.ClassName}}ServiceImpl extends ServiceImpl<{{.ClassName}}Mapper, {{.ClassName}}> implements I{{.ClassName}}Service {

}
