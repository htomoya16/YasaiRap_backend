env "local" {
    # 完成図（schema/配下の .sql/.hcl）
    src = "file://schema"
    dev = "docker://postgres/16/yasairap"
    
    # マイグレーション履歴の置き場
    migration {
        dir = "file://migrations"
        format = atlas
    }
    # 実際に適用する接続先
    url = "postgres://yasairap_user:yasairap_password@localhost:5432/yasairap?sslmode=disable"
    exclude = [
        "atlas_schema_revisions",
    ]
}