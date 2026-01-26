package com.storytrim.app.ui.points.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.recyclerview.widget.DiffUtil
import androidx.recyclerview.widget.ListAdapter
import androidx.recyclerview.widget.RecyclerView
import com.storytrim.app.data.dto.PointsLedgerItem
import com.storytrim.app.databinding.ItemPointsLedgerBinding

class PointsLedgerAdapter : ListAdapter<PointsLedgerItem, PointsLedgerAdapter.LedgerViewHolder>(DiffCallback()) {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): LedgerViewHolder {
        val binding = ItemPointsLedgerBinding.inflate(LayoutInflater.from(parent.context), parent, false)
        return LedgerViewHolder(binding)
    }

    override fun onBindViewHolder(holder: LedgerViewHolder, position: Int) {
        holder.bind(getItem(position))
    }

    class LedgerViewHolder(private val binding: ItemPointsLedgerBinding) : RecyclerView.ViewHolder(binding.root) {
        fun bind(item: PointsLedgerItem) {
            binding.tvTitle.text = mapTitle(item.reason)
            binding.tvSubtitle.text = buildSubtitle(item.extra)
            binding.tvSubtitle.visibility = if (binding.tvSubtitle.text.isNullOrBlank()) View.GONE else View.VISIBLE
            binding.tvTime.text = item.createdAt
            val deltaText = if (item.change > 0) "+${item.change}" else item.change.toString()
            binding.tvDelta.text = deltaText
            val color = if (item.change > 0) {
                binding.root.context.getColor(com.storytrim.app.R.color.teal_600)
            } else {
                binding.root.context.getColor(com.storytrim.app.R.color.stone_700)
            }
            binding.tvDelta.setTextColor(color)
        }

        private fun mapTitle(reason: String): String {
            return when (reason) {
                "register_bonus" -> "注册赠送"
                "trim_use" -> "精简消耗"
                "trim_refund" -> "精简退款"
                "recharge" -> "积分充值"
                "manual_adjust" -> "积分调整"
                else -> "积分变动"
            }
        }

        private fun buildSubtitle(extra: Map<String, String>?): String {
            if (extra.isNullOrEmpty()) return ""
            val parts = mutableListOf<String>()
            extra["book_title"]?.let { parts.add("《$it》") }
            extra["chapter_title"]?.let { parts.add(it) }
            extra["prompt_name"]?.let { parts.add(it) }
            return parts.joinToString(" · ")
        }
    }

    private class DiffCallback : DiffUtil.ItemCallback<PointsLedgerItem>() {
        override fun areItemsTheSame(oldItem: PointsLedgerItem, newItem: PointsLedgerItem): Boolean {
            return oldItem.id == newItem.id
        }

        override fun areContentsTheSame(oldItem: PointsLedgerItem, newItem: PointsLedgerItem): Boolean {
            return oldItem == newItem
        }
    }
}
