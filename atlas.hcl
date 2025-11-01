env "local" {
    # 完成図（schema/配下の .sql/.hcl）
    src = "file://schema"
    # 差分計画用の一時DB
    dev-url = "docker://mysql/8.4"
    
    # マイグレーション履歴の置き場
    migration {
        dir = "file://migrations"
        format = atlas
    }
    # 実際に適用する接続先
    url = "mysql://yasairap_user:yasairap_password@localhost:3306/yasairap"
}