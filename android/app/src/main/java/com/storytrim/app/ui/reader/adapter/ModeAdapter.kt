package com.storytrim.app.ui.reader.adapter

import android.graphics.Color
import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.RecyclerView
import com.storytrim.app.data.model.Prompt
import com.storytrim.app.databinding.ItemModeGridBinding

class ModeAdapter(
    private val prompts: List<Prompt>,
    currentModeId: Int,
    private val onModeClick: (Int) -> Unit
) : RecyclerView.Adapter<ModeAdapter.ModeViewHolder>() {

    companion object {
        private val COLOR_DEFAULT = Color.parseColor("#FFFFFF")
        private val COLOR_SELECTED = Color.parseColor("#F0FDF4")
        private val COLOR_BORDER_DEFAULT = Color.parseColor("#E5E7EB")
        private val COLOR_BORDER_SELECTED = Color.parseColor("#14B8A6")
        private val COLOR_TEXT_DEFAULT = Color.parseColor("#374151")
        private val COLOR_TEXT_SELECTED = Color.parseColor("#0F766E")
    }

    private var currentModeId: Int = currentModeId
        set(value) {
            field = value
            notifyDataSetChanged()
        }

    fun updateSelectedId(newId: Int) {
        val oldId = currentModeId
        currentModeId = newId

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
        holder.bind(prompts[position])
    }

    override fun getItemCount() = prompts.size

    inner class ModeViewHolder(private val binding: ItemModeGridBinding) :
        RecyclerView.ViewHolder(binding.root) {

        fun bind(prompt: Prompt) {
            binding.tvModeName.text = prompt.name

            val isSelected = prompt.id == currentModeId
            if (isSelected) {
                binding.tvModeName.setTextColor(COLOR_TEXT_SELECTED)
                binding.root.setStrokeColor(COLOR_BORDER_SELECTED)
                binding.root.setCardBackgroundColor(COLOR_SELECTED)
            } else {
                binding.tvModeName.setTextColor(COLOR_TEXT_DEFAULT)
                binding.root.setStrokeColor(COLOR_BORDER_DEFAULT)
                binding.root.setCardBackgroundColor(COLOR_DEFAULT)
            }

            binding.root.setOnClickListener {
                onModeClick(prompt.id)
            }
        }
    }
}
