server:
  port: 8080
  servlet:
    context-path: /

spring:
  application:
    name: {{.Config.PackageConfig.BasePackage | replace "." "-"}}-service
  
  # 数据源配置
  datasource:
    driver-class-name: com.mysql.cj.jdbc.Driver
    url: jdbc:mysql://localhost:3306/{{.Config.PackageConfig.BasePackage | replace "." "_"}}?useUnicode=true&characterEncoding=utf8&zeroDateTimeBehavior=convertToNull&useSSL=true&serverTimezone=GMT%2B8
    username: root
    password: 123456
    
  # JPA配置
  jpa:
    hibernate:
      ddl-auto: update
    show-sql: true
    database-platform: org.hibernate.dialect.MySQL8Dialect

# MyBatis Plus配置
mybatis-plus:
  configuration:
    # 开启驼峰命名转换
    map-underscore-to-camel-case: true
    # 开启SQL日志
    log-impl: org.apache.ibatis.logging.stdout.StdOutImpl
  global-config:
    db-config:
      # 主键策略
      id-type: auto
      # 逻辑删除字段
      logic-delete-field: deleted
      # 逻辑删除值
      logic-delete-value: 1
      # 逻辑未删除值
      logic-not-delete-value: 0
  # Mapper XML文件位置
  mapper-locations: classpath*:mapper/*.xml
  # 实体类包路径
  type-aliases-package: {{.Config.PackageConfig.EntityPackage}}

# 日志配置
logging:
  level:
    {{.Config.PackageConfig.BasePackage}}: debug
    org.springframework.web: info
  pattern:
    console: "%d{yyyy-MM-dd HH:mm:ss} [%thread] %-5level %logger{36} - %msg%n"
    file: "%d{yyyy-MM-dd HH:mm:ss} [%thread] %-5level %logger{36} - %msg%n"
  file:
    name: logs/application.log

# 管理端点配置
management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics
  endpoint:
    health:
      show-details: always