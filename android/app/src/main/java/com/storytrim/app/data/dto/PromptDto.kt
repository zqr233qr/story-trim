package com.storytrim.app.data.dto

import com.google.gson.annotations.SerializedName

data class PromptRespDto(
    @SerializedName("id") val id: Int,
    @SerializedName("name") val name: String,
    @SerializedName("description") val description: String,
    @SerializedName("is_default") val is_default: Boolean
)
