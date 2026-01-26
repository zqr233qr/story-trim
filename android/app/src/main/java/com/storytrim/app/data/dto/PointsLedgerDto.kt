package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class PointsLedgerResp(
    @SerializedName("items") val items: List<PointsLedgerItem> = emptyList()
)

data class PointsLedgerItem(
    @SerializedName("id") val id: Long,
    @SerializedName("change") val change: Int,
    @SerializedName("balance_after") val balanceAfter: Int,
    @SerializedName("type") val type: String,
    @SerializedName("reason") val reason: String,
    @SerializedName("ref_type") val refType: String? = null,
    @SerializedName("ref_id") val refId: String? = null,
    @SerializedName("extra") val extra: Map<String, String>? = null,
    @SerializedName("created_at") val createdAt: String
)
