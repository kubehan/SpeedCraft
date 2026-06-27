-- Seed data for SpeedCraft
-- Run: sqlite3 data/speedcraft.db < scripts/seed.sql

INSERT OR IGNORE INTO projects (title, slug, summary, content, category, tech_stack, client_name, is_published) VALUES
('电商平台K8s迁移',
 'ecommerce-k8s-migration',
 '为某头部电商平台完成从传统VM架构到Kubernetes容器平台的全面迁移，实现弹性伸缩和自动化运维。',
 '<h2>项目背景</h2><p>该电商平台日活用户超过100万，原有VM架构面临扩容慢、资源利用率低等问题。</p><h2>解决方案</h2><ul><li>Kubernetes集群设计与搭建</li><li>微服务容器化改造</li><li>CI/CD流水线建设</li><li>Prometheus+Grafana监控体系</li></ul><h2>成果</h2><p>部署效率提升10倍，资源利用率提高60%，3次大促零宕机。</p>',
 'Kubernetes',
 'Docker,Kubernetes,Helm,Terraform,Prometheus',
 '某电商平台',
 1),

('CI/CD流水线重构',
 'cicd-pipeline-redesign',
 '为金融科技公司重构CI/CD流水线，实现多环境自动部署与安全合规门禁。',
 '<h2>项目背景</h2><p>原有Jenkins流水线维护困难，部署需要手动操作，无法满足合规审计要求。</p><h2>解决方案</h2><ul><li>GitLab CI + GitOps 工作流</li><li>多环境（dev/staging/prod）自动部署</li><li>安全扫描与合规检查集成</li><li>灰度发布与自动回滚</li></ul><h2>成果</h2><p>发布频率从周级降至日级，回滚时间从30分钟降至2分钟。</p>',
 'CI/CD',
 'GitLab CI,ArgoCD,Terraform,Trivy,Opa',
 '某金融科技公司',
 1),

('全链路监控平台建设',
 'full-stack-monitoring',
 '为物联网企业构建从基础设施到业务指标的全链路监控平台。',
 '<h2>项目背景</h2><p>设备量激增至50万台，原有监控系统无法承载，故障定位困难。</p><h2>解决方案</h2><ul><li>Prometheus联邦集群</li><li>Grafana统一看板</li><li>ELK日志平台</li><li>OpenTelemetry链路追踪</li><li>AlertManager告警值班</li></ul>',
 '监控',
 'Prometheus,Grafana,ELK,OpenTelemetry',
 '某IoT企业',
 1);

INSERT OR IGNORE INTO blog_posts (title, slug, summary, content, tags, is_published) VALUES
('9年运维经验总结：如何构建高可用架构',
 'high-availability-architecture',
 '本文总结了我在9年运维生涯中关于高可用架构设计的关键原则和实践经验，包括冗余设计、故障隔离、容量规划等核心话题。',
 '<h2>引言</h2><p>高可用架构是每个运维工程师的必修课。经过9年的实战积累，我想分享一些核心原则。</p><h2>1. 冗余设计</h2><p>消除单点故障是高可用的第一步。从网络层到应用层，每一层都需要冗余设计。</p><h2>2. 故障隔离</h2><p>使用熔断器、舱壁模式等技术防止故障级联扩散。</p><h2>3. 容量规划</h2><p>基于历史数据和业务预测进行容量规划，预留20%-30%的Buffer。</p><h2>4. 自动化运维</h2><p>所有重复性工作都应该自动化，减少人为失误。</p><h2>总结</h2><p>高可用不是一蹴而就的，需要持续迭代和改进。</p>',
 '高可用,架构设计,SRE',
 1),

('Kubernetes排障实战指南',
 'kubernetes-troubleshooting-guide',
 '整理了日常K8s运维中常见的故障场景和排障方法，包含Pod异常、网络问题、存储故障等实战案例。',
 '<h2>Pod异常排障</h2><p>CrashLoopBackOff、ImagePullBackOff、Pending状态的排查步骤...</p><h2>网络问题</h2><p>DNS解析异常、Service访问超时、Ingress配置错误的排查...</p><h2>存储故障</h2><p>PV/PVC绑定失败、存储性能问题的排查...</p>',
 'Kubernetes,排障,运维',
 1),

('Terraform最佳实践',
 'terraform-best-practices',
 '分享团队在使用Terraform管理基础设施过程中总结的最佳实践，包括项目结构、状态管理、模块设计等。',
 '<h2>项目结构</h2><p>推荐的分层目录结构，环境隔离策略...</p><h2>状态管理</h2><p>远程状态存储、锁机制、状态迁移...</p><h2>模块设计</h2><p>可复用模块的设计原则、版本管理...</p>',
 'Terraform,IaC,最佳实践',
 1);
