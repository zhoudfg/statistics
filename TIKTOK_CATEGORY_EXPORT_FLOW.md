# TikTok 类目拉取、清洗及导出流程

本文档说明从 TikTok 拉取类目，到最终写入 `temp1_tiktok_category_export` 的完整流程。所有命令均在项目根目录执行：

```powershell
Set-Location E:\phpstudy_pro\WWW\statistics
```

## 1. 流程总览

```text
TikTok 类目 API
  -> category_sync
  -> temp1_tiktok_category
  -> category_path_clean
  -> temp1_tiktok_category_path
  -> category_attr_sync（TikTok 类目属性 API）
  -> temp1_tiktok_category_attr
  -> category_attr_clean
  -> temp1_tiktok_category_export
```

店铺类型：

- `shop_type=1`：全托管，`seller_type=ALL`。
- `shop_type=2`：半托管，需区分 `version=v1/v2` 和 `seller_type=LOCAL/CROSS_BORDER`。

运行前确认：

1. 已安装 Go（项目要求 Go 1.17 或兼容版本）。
2. `statistics-server/server.yaml` 中 `mysql.erp` 可连接，目标库中已存在相关表。
3. `statistics-server/server.yaml` 中 TikTok 各店铺的凭证、区域及版本配置有效。
4. TikTok 凭证属于敏感信息，禁止将配置内容或执行日志中的凭证粘贴到文档、工单或群聊。

可先执行测试：

```powershell
go test ./statistics-server/tiktok/category_sync ./statistics-server/tiktok/category_path_clean ./statistics-server/tiktok/category_attr_sync ./statistics-server/tiktok/category_attr_clean
```

## 2. 第一步：拉取类目到 temp1_tiktok_category

### 2.1 全托管类目

先试跑，只请求和解析，不写数据库：

```powershell
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 1 -page-size 50 -max-pages 0 -db-batch-size 500 -request-timeout 180s -dry-run -verbose
```

正式执行：

```powershell
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 1 -page-size 50 -max-pages 0 -db-batch-size 500 -request-timeout 180s -verbose
```

全托管接口分页拉取；接口返回的 `version` 会写入表中。若只想保留某版本，可额外添加 `-version v1` 或 `-version v2`。

### 2.2 半托管类目

半托管接口每次只返回一种语言，因此每个“版本 + 卖家类型”必须分别执行中文和英文两次。Upsert 时中文只更新 `name_cn`，英文只更新 `name_en`。

LOCAL v1：

```powershell
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 2 -version v1 -seller-type LOCAL -locale zh-CN -db-batch-size 500 -request-timeout 180s -verbose
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 2 -version v1 -seller-type LOCAL -locale en-US -db-batch-size 500 -request-timeout 180s -verbose
```

LOCAL v2：

```powershell
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 2 -version v2 -seller-type LOCAL -locale zh-CN -db-batch-size 500 -request-timeout 180s -verbose
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 2 -version v2 -seller-type LOCAL -locale en-US -db-batch-size 500 -request-timeout 180s -verbose
```

CROSS_BORDER v1：

```powershell
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 2 -version v1 -seller-type CROSS_BORDER -locale zh-CN -db-batch-size 500 -request-timeout 180s -verbose
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 2 -version v1 -seller-type CROSS_BORDER -locale en-US -db-batch-size 500 -request-timeout 180s -verbose
```

CROSS_BORDER v2：

```powershell
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 2 -version v2 -seller-type CROSS_BORDER -locale zh-CN -db-batch-size 500 -request-timeout 180s -verbose
go run ./statistics-server/tiktok/category_sync -config-dir ./statistics-server -shop-type 2 -version v2 -seller-type CROSS_BORDER -locale en-US -db-batch-size 500 -request-timeout 180s -verbose
```

试跑任一半托管组合时，在命令末尾添加 `-dry-run`。

### 2.3 category_sync 全部参数

| 参数 | 默认值 | 说明 |
|---|---:|---|
| `-config-dir` | 自动识别 | `server.yaml` 所在目录；推荐显式传 `./statistics-server`。 |
| `-shop-type` | `0` | 必填；`1=full`，`2=half`。 |
| `-version` | 空 | `v1` 或 `v2`；半托管必填，全托管可选。 |
| `-seller-type` | 空 | 半托管必填：`LOCAL` 或 `CROSS_BORDER`；全托管仅允许空或 `ALL`。 |
| `-locale` | `zh-CN` | 半托管语言：`zh-CN` 或 `en-US`。 |
| `-page-size` | `50` | 全托管分页大小，范围 `1~50`。 |
| `-max-pages` | `0` | 全托管最多请求页数；`0` 表示不限。 |
| `-db-batch-size` | `500` | 每批 Upsert 行数，必须大于 0。 |
| `-request-timeout` | `180s` | 每次 TikTok 请求超时，Go duration 格式，如 `30s`、`3m`。 |
| `-dry-run` | `false` | 调接口并解析，但不写 `temp1_tiktok_category`。 |
| `-verbose` | `false` | 输出分页及明细日志。 |

## 3. 第二步：清洗叶子类目路径到 temp1_tiktok_category_path

该步骤读取 `temp1_tiktok_category`，按 `(shop_type, version, seller_type)` 分组，从每个叶子节点向上查找父节点，生成中英文名称路径和 ID 路径。

先对全量数据试跑：

```powershell
go run ./statistics-server/tiktok/category_path_clean -config-dir ./statistics-server -shop-type 0 -separator " > " -status 0 -db-batch-size 500 -skip-missing-parent -dry-run -verbose
```

确认 `missing_parent` 和 `cycle` 日志后正式执行：

```powershell
go run ./statistics-server/tiktok/category_path_clean -config-dir ./statistics-server -shop-type 0 -separator " > " -status 0 -db-batch-size 500 -skip-missing-parent -verbose
```

`-status 0` 表示路径等待属性拉取。建议保留 `-skip-missing-parent`，避免把父链不完整的路径写入后续流程；循环父链始终会跳过。

如需单独重建某个范围：

```powershell
go run ./statistics-server/tiktok/category_path_clean -config-dir ./statistics-server -shop-type 1 -seller-type ALL -separator " > " -status 0 -db-batch-size 500 -skip-missing-parent -verbose
go run ./statistics-server/tiktok/category_path_clean -config-dir ./statistics-server -shop-type 2 -version v2 -seller-type LOCAL -separator " > " -status 0 -db-batch-size 500 -skip-missing-parent -verbose
```

### category_path_clean 全部参数

| 参数 | 默认值 | 说明 |
|---|---:|---|
| `-config-dir` | 自动识别 | `server.yaml` 所在目录。 |
| `-shop-type` | `0` | `0=全部`，`1=全托管`，`2=半托管`。 |
| `-version` | 空 | 空=全部，也可传 `v1`、`v2`。 |
| `-seller-type` | 空 | 空=全部，也可传 `LOCAL`、`CROSS_BORDER`、`ALL`。 |
| `-separator` | ` > ` | 路径分隔符，不可为空。PowerShell 中建议加双引号。 |
| `-status` | `0` | 写入路径表的状态：`0=待处理`、`1=成功`、`2=失败`。正常新流程使用 0。 |
| `-db-batch-size` | `500` | 每批 Upsert 行数。 |
| `-skip-missing-parent` | `false` | 跳过父链不完整的叶子类目。 |
| `-dry-run` | `false` | 只构建和统计，不写路径表。 |
| `-verbose` | `false` | 输出缺失父节点及循环路径明细。 |

## 4. 第三步：拉取类目属性到 temp1_tiktok_category_attr

属性任务读取路径表中 `status=0` 的叶子类目，并发请求 TikTok 属性接口。成功后 Upsert 原始 JSON 到 `temp1_tiktok_category_attr`，并把路径状态改为 `1`；失败则改为 `2`。

### 4.1 全托管

试跑：

```powershell
go run ./statistics-server/tiktok/category_attr_sync -config-dir ./statistics-server -shop-type 1 -status 0 -workers 8 -batch-size 500 -limit 0 -retries 2 -retry-wait 2s -request-timeout 180s -dry-run -verbose
```

正式执行：

```powershell
go run ./statistics-server/tiktok/category_attr_sync -config-dir ./statistics-server -shop-type 1 -status 0 -workers 8 -batch-size 500 -limit 0 -retries 2 -retry-wait 2s -request-timeout 180s -verbose
```

### 4.2 半托管四种组合

```powershell
go run ./statistics-server/tiktok/category_attr_sync -config-dir ./statistics-server -shop-type 2 -status 0 -version v1 -seller-type LOCAL -workers 8 -batch-size 500 -limit 0 -retries 2 -retry-wait 2s -request-timeout 180s -verbose
go run ./statistics-server/tiktok/category_attr_sync -config-dir ./statistics-server -shop-type 2 -status 0 -version v2 -seller-type LOCAL -workers 8 -batch-size 500 -limit 0 -retries 2 -retry-wait 2s -request-timeout 180s -verbose
go run ./statistics-server/tiktok/category_attr_sync -config-dir ./statistics-server -shop-type 2 -status 0 -version v1 -seller-type CROSS_BORDER -workers 8 -batch-size 500 -limit 0 -retries 2 -retry-wait 2s -request-timeout 180s -verbose
go run ./statistics-server/tiktok/category_attr_sync -config-dir ./statistics-server -shop-type 2 -status 0 -version v2 -seller-type CROSS_BORDER -workers 8 -batch-size 500 -limit 0 -retries 2 -retry-wait 2s -request-timeout 180s -verbose
```

失败重试：将 `-status 0` 改为 `-status 2`，其余筛选参数保持一致，例如：

```powershell
go run ./statistics-server/tiktok/category_attr_sync -config-dir ./statistics-server -shop-type 2 -status 2 -version v2 -seller-type LOCAL -workers 4 -batch-size 200 -limit 0 -retries 3 -retry-wait 3s -request-timeout 180s -verbose
```

### category_attr_sync 全部参数

| 参数 | 默认值 | 说明 |
|---|---:|---|
| `-config-dir` | 自动识别 | `server.yaml` 所在目录。 |
| `-shop-type` | `0` | `0=全部`，`1=全托管`，`2=半托管`。半托管必须同时指定版本和卖家类型。 |
| `-status` | `0` | 只允许 `0=待处理` 或 `2=失败重试`。 |
| `-version` | 空 | 空=全部，或 `v1`、`v2`、`null`。半托管必须为 v1/v2。 |
| `-seller-type` | 空 | `LOCAL` 或 `CROSS_BORDER`，且仅能与 `shop-type=2` 一起使用。 |
| `-workers` | `8` | TikTok API 并发数，必须大于 0。遇到限流时调低。 |
| `-batch-size` | `500` | 每批读取路径行数。 |
| `-limit` | `0` | 最多入队行数；`0` 表示不限。 |
| `-skip-existing` | `false` | 跳过属性表中已存在相同唯一键的路径。 |
| `-retries` | `2` | 单次请求失败后的重试次数；总尝试次数为 `retries + 1`。 |
| `-retry-wait` | `2s` | 线性退避基数；第 n 次等待 `retry-wait × n`。 |
| `-request-timeout` | `180s` | 每次 API 请求超时。 |
| `-dry-run` | `false` | 调接口并验证 JSON，但不写属性表，也不更新路径状态。 |
| `-verbose` | `false` | 输出每个类目的请求开始和结束日志。 |

## 5. 第四步：清洗属性并写入 temp1_tiktok_category_export

该步骤把属性原始 JSON 与路径表按 `(category_id, shop_type, version, seller_type)` 关联；一个类目属性会生成一行 export 数据。成功写入后，将对应 `temp1_tiktok_category_attr.is_clean` 更新为 `1`。每个属性源记录在独立事务中完成“写 export + 标记已清洗”。

先试跑全量未清洗数据：

```powershell
go run ./statistics-server/tiktok/category_attr_clean -config-dir ./statistics-server -shop-type 0 -workers 8 -batch-size 500 -limit 0 -after-id 0 -dry-run -verbose
```

首次正式执行：

```powershell
go run ./statistics-server/tiktok/category_attr_clean -config-dir ./statistics-server -shop-type 0 -workers 8 -batch-size 500 -limit 0 -after-id 0 -verbose
```

默认只处理 `is_clean IS NULL OR is_clean=0` 的属性行。若需要重新清洗已处理数据，必须同时启用 `-include-cleaned` 和 `-replace-existing`，先软删除相同路径的旧 export 行，再插入新行：

```powershell
go run ./statistics-server/tiktok/category_attr_clean -config-dir ./statistics-server -shop-type 2 -version v2 -seller-type LOCAL -workers 8 -batch-size 500 -limit 0 -after-id 0 -include-cleaned -replace-existing -verbose
```

中断后可根据上一次日志中的 `last_attr_id` 断点续跑，例如：

```powershell
go run ./statistics-server/tiktok/category_attr_clean -config-dir ./statistics-server -shop-type 0 -workers 8 -batch-size 500 -limit 0 -after-id 123456 -verbose
```

注意：首次执行不要使用 `-replace-existing`；普通续跑也不需要 `-include-cleaned`。单独使用 `-include-cleaned` 而不使用 `-replace-existing`，可能重复插入 export 数据。

### category_attr_clean 全部参数

| 参数 | 默认值 | 说明 |
|---|---:|---|
| `-config-dir` | 自动识别 | `server.yaml` 所在目录。 |
| `-shop-type` | `0` | `0=全部`，`1=全托管`，`2=半托管`。 |
| `-version` | 空 | 空=全部，或 `v1`、`v2`、`null`。 |
| `-seller-type` | 空 | 空=全部，或 `LOCAL`、`CROSS_BORDER`、`ALL`、`null`。 |
| `-workers` | `8` | 并发数据库清洗工作线程数。 |
| `-batch-size` | `500` | 每批读取的属性源记录数。 |
| `-limit` | `0` | 本次最多清洗的属性源记录数；`0` 表示不限。 |
| `-after-id` | `0` | 只处理 `temp1_tiktok_category_attr.id` 大于该值的记录，用于断点续跑。 |
| `-include-cleaned` | `false` | 同时处理 `is_clean=1` 的记录。仅重建时使用。 |
| `-replace-existing` | `false` | 插入前软删除相同店铺、版本、卖家及路径的现有 export 行。 |
| `-dry-run` | `false` | 解析并计数，不写 export，也不更新 `is_clean`。 |
| `-verbose` | `false` | 输出每条属性源记录的清洗日志。 |

## 6. 数据核对

以下 SQL 可在 MySQL 客户端执行。连接参数应使用实际环境值，不要把密码直接写入共享脚本：

```powershell
mysql --host=127.0.0.1 --port=3306 --user=<用户名> --password --database=erp
```

进入 MySQL 后执行：

```sql
-- 类目数量及中英文名称覆盖情况
SELECT shop_type, version, seller_type,
       COUNT(*) AS category_count,
       SUM(is_leaf = 1) AS leaf_count,
       SUM(name_cn IS NULL OR name_cn = '') AS missing_cn,
       SUM(name_en IS NULL OR name_en = '') AS missing_en
FROM temp1_tiktok_category
WHERE deleted_at IS NULL
GROUP BY shop_type, version, seller_type;

-- 路径状态：0 待拉属性，1 成功，2 失败
SELECT shop_type, version, seller_type, status, COUNT(*) AS path_count
FROM temp1_tiktok_category_path
WHERE deleted_at IS NULL
GROUP BY shop_type, version, seller_type, status;

-- 属性原始数据及清洗状态
SELECT shop_type, version, seller_type, COALESCE(is_clean, 0) AS is_clean,
       COUNT(*) AS attr_source_count
FROM temp1_tiktok_category_attr
WHERE deleted_at IS NULL
GROUP BY shop_type, version, seller_type, COALESCE(is_clean, 0);

-- 最终导出数据量
SELECT shop_type, version, seller_type, COUNT(*) AS export_count
FROM temp1_tiktok_category_export
WHERE deleted_at IS NULL
GROUP BY shop_type, version, seller_type;

-- 找出成功路径中缺失属性原始记录的类目
SELECT p.id, p.category_id, p.shop_type, p.version, p.seller_type
FROM temp1_tiktok_category_path p
LEFT JOIN temp1_tiktok_category_attr a
  ON a.category_id = p.category_id
 AND a.shop_type <=> p.shop_type
 AND a.version <=> p.version
 AND a.seller_type <=> p.seller_type
 AND a.deleted_at IS NULL
WHERE p.deleted_at IS NULL
  AND p.status = 1
  AND a.id IS NULL;
```

## 7. 推荐生产执行顺序

1. 执行第 2 节类目拉取：全托管 1 次，半托管 4 个组合各执行中英文 2 次。
2. 执行 `category_path_clean` 的 `-dry-run`，检查缺失父节点和循环。
3. 正式执行 `category_path_clean`，路径状态写 0。
4. 分别执行全托管及四种半托管 `category_attr_sync`。
5. 对失败路径使用 `-status 2` 重试，直到失败量可接受或已定位原因。
6. 执行 `category_attr_clean -dry-run`，核对预计清洗量和导出行数。
7. 正式执行 `category_attr_clean`。
8. 执行第 6 节 SQL，对类目、路径、属性源及 export 四层数据量进行核对。

