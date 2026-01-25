package com.storytrim.app.ui.reader.adapter

import android.graphics.Color
import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.storytrim.app.data.model.Prompt
import com.storytrim.app.databinding.ItemAiTrimOptionBinding

class AiTrimOptionAdapter(
    private val prompts: List<Prompt>,
    private var selectedId: Int,
    private var trimmedPromptIds: Set<Int>,
    private val onSelect: (Int) -> Unit
) : RecyclerView.Adapter<AiTrimOptionAdapter.OptionViewHolder>() {

    fun updateSelectedId(newId: Int) {
        val oldId = selectedId
        selectedId = newId
        val oldIndex = prompts.indexOfFirst { it.id == oldId }
        val newIndex = prompts.indexOfFirst { it.id == newId }
        if (oldIndex >= 0) notifyItemChanged(oldIndex)
        if (newIndex >= 0) notifyItemChanged(newIndex)
    }

    fun updateTrimmedIds(newIds: Set<Int>) {
        trimmedPromptIds = newIds
        notifyDataSetChanged()
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): OptionViewHolder {
        val binding = ItemAiTrimOptionBinding.inflate(
            LayoutInflater.from(parent.context), parent, false
        )
        return OptionViewHolder(binding)
    }

    override fun onBindViewHolder(holder: OptionViewHolder, position: Int) {
        holder.bind(prompts[position], selectedId, trimmedPromptIds, onSelect)
    }

    override fun getItemCount(): Int = prompts.size

    inner class OptionViewHolder(private val binding: ItemAiTrimOptionBinding) :
        RecyclerView.ViewHolder(binding.root) {

        fun bind(prompt: Prompt, selectedId: Int, trimmedPromptIds: Set<Int>, onSelect: (Int) -> Unit) {
            val isSelected = prompt.id == selectedId
            val isTrimmed = trimmedPromptIds.contains(prompt.id)
            binding.tvTitle.text = prompt.name
            binding.tvDesc.text = prompt.description

            if (isTrimmed) {
                binding.tvTag.visibility = android.view.View.VISIBLE
                binding.tvTag.text = "已精简"
            } else {
                binding.tvTag.visibility = android.view.View.GONE
            }

            if (isSelected) {
                binding.card.setCardBackgroundColor(Color.parseColor("#ECFDFB"))
                binding.card.strokeColor = Color.parseColor("#14B8A6")
                binding.tvTitle.setTextColor(Color.parseColor("#0F766E"))
                binding.tvDesc.setTextColor(Color.parseColor("#0F766E"))
            } else {
                binding.card.setCardBackgroundColor(Color.parseColor("#FFFFFF"))
                binding.card.strokeColor = Color.parseColor("#E7E5E4")
                binding.tvTitle.setTextColor(Color.parseColor("#1C1917"))
                binding.tvDesc.setTextColor(Color.parseColor("#78716C"))
            }

            binding.root.setOnClickListener {
                onSelect(prompt.id)
            }
        }
    }
}
