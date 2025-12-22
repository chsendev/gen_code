package {{.ControllerPackage}};

import org.springframework.web.bind.annotation.*;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import org.springframework.beans.factory.annotation.Autowired;
import java.util.List;
import {{.EntityPackage}}.{{.ClassName}};
import {{.ServicePackage}}.I{{.ClassName}}Service;

/**
 * {{.Table.TableComment}}Controller
 * @author {{.Author}}
 * @date {{.Date}}
 */
@RestController
@RequestMapping("/{{.Table.TableName}}")
public class {{.ClassName}}Controller {

    @Autowired
    private I{{.ClassName}}Service {{.Table.TableName}}Service;

    /**
     * 查询{{.Table.TableComment}}列表
     */
    @GetMapping("/list")
    public List<{{.ClassName}}> list({{.ClassName}} {{.Table.TableName}}) {
        return {{.Table.TableName}}Service.list();
    }

    /**
     * 查询{{.Table.TableComment}}分页列表
     */
    @GetMapping("/page")
    public Page<{{.ClassName}}> page(Page<{{.ClassName}}> page, {{.ClassName}} {{.Table.TableName}}) {
        return {{.Table.TableName}}Service.page(page);
    }

    /**
     * 获取{{.Table.TableComment}}详细信息
     */
    @GetMapping("/{id}")
    public {{.ClassName}} getInfo(@PathVariable("id") {{.Table.PrimaryKey.JavaType}} id) {
        return {{.Table.TableName}}Service.getById(id);
    }

    /**
     * 新增{{.Table.TableComment}}
     */
    @PostMapping
    public boolean add(@RequestBody {{.ClassName}} {{.Table.TableName}}) {
        return {{.Table.TableName}}Service.save({{.Table.TableName}});
    }

    /**
     * 修改{{.Table.TableComment}}
     */
    @PutMapping
    public boolean edit(@RequestBody {{.ClassName}} {{.Table.TableName}}) {
        return {{.Table.TableName}}Service.updateById({{.Table.TableName}});
    }

    /**
     * 删除{{.Table.TableComment}}
     */
    @DeleteMapping("/{id}")
    public boolean delete(@PathVariable("id") {{.Table.PrimaryKey.JavaType}} id) {
        return {{.Table.TableName}}Service.removeById(id);
    }

}
