@@Meta.Output="/src/main/java/{{.Config.PackageConfig.BasePackage | replace "." "/"}}/Application.java"

package {{.Config.PackageConfig.BasePackage}};

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * SpringBoot启动类
 * 
 * @author {{.Config.GenConfig.Author}}
 * @date {{.Config.GenConfig.Date}}
 */
@SpringBootApplication
@MapperScan("{{.Config.PackageConfig.MapperPackage}}")
public class Application {

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
        System.out.println("==========================================");
        System.out.println("应用启动成功！");
        System.out.println("==========================================");
    }
}