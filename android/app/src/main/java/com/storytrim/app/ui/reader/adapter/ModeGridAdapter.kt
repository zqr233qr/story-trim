package com.storytrim.app.ui.reader.adapter

import android.graphics.Color
import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.storytrim.app.data.model.Prompt
import com.storytrim.app.databinding.ItemModeGridBinding

class ModeGridAdapter(
    private val prompts: List<Prompt>,
    private var selectedPromptId: Int,
    private val onPromptClick: (Int) -> Unit
) : RecyclerView.Adapter<ModeGridAdapter.ModeViewHolder>() {

    fun updateSelectedId(newId: Int) {
        val oldId = selectedPromptId
        selectedPromptId = newId

        val oldPos = prompts.indexOfFirst { it.id == oldId }
        val newPos = prompts.indexOfFirst { it.id == newId }

        if (oldPos >= 0) notifyItemChanged(oldPos)
        if (newPos >= 0) notifyItemChanged(newPos)
    }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): ModeViewHolder {
        val binding = ItemModeGridBinding.inflate(
            LayoutInflater.from(parent.context), parent, false
        )
        return ModeViewHolder(binding)
    }

    override fun onBindViewHolder(holder: ModeViewHolder, position: Int) {
        holder.bind(prompts[position], selectedPromptId, onPromptClick)
    }

    override fun getItemCount() = prompts.size

    inner class ModeViewHolder(private val binding: ItemModeGridBinding) :
        RecyclerView.ViewHolder(binding.root) {

        fun bind(prompt: Prompt, selectedId: Int, onClick: (Int) -> Unit) {
            binding.tvModeName.text = prompt.name

            val isSelected = prompt.id == selectedId
            if (isSelected) {
                binding.tvModeName.setTextColor(Color.parseColor("#0F766E"))
                binding.root.setCardBackgroundColor(Color.parseColor("#F0FDFA"))
                binding.root.setStrokeColor(Color.parseColor("#14B8A6"))
            } else {
                binding.tvModeName.setTextColor(Color.parseColor("#78716C"))
                binding.root.setCardBackgroundColor(Color.parseColor("#FAF9F6"))
                binding.root.setStrokeColor(Color.parseColor("#E7E5E4"))
            }

            binding.root.setOnClickListener {
                onClick(prompt.id)
            }
        }
    }
}
