package com.storytrim.app.ui.reader.adapter

import android.graphics.Color
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.recyclerview.widget.DiffUtil
import androidx.recyclerview.widget.ListAdapter
import androidx.recyclerview.widget.RecyclerView
import com.storytrim.app.data.dto.ChapterTrimOption
import com.storytrim.app.data.dto.TrimStatus
import com.storytrim.app.databinding.ItemChapterTrimBinding

class ChapterTrimAdapter(
    private val onChapterClick: (ChapterTrimOption) -> Unit,
    private val currentChapterId: Long
) : ListAdapter<ChapterTrimOption, ChapterTrimAdapter.ChapterTrimViewHolder>(ChapterTrimDiffCallback()) {

    companion object {
        private val COLOR_DEFAULT = Color.parseColor("#FFFFFF")
        private val COLOR_SELECTED = Color.parseColor("#ECFDFB")
        private val COLOR_DISABLED = Color.parseColor("#F5F5F4")
        private val COLOR_BORDER_DEFAULT = Color.parseColor("#E7E5E4")
        private val COLOR_BORDER_SELECTED = Color.parseColor("#14B8A6")
        private val COLOR_TEXT_DEFAULT = Color.parseColor("#1C1917")
        private val COLOR_TEXT_DISABLED = Color.parseColor("#9CA3AF")
        private val COLOR_TEXT_SELECTED = Color.parseColor("#0F766E")
    }

    var selectedChapterIds = setOf<Long>()
        set(value) {
            val oldIds = field
            field = value

            val added = value - oldIds
            val removed = oldIds - value

            currentList.forEachIndexed { index, option ->
                if (added.contains(option.id) || removed.contains(option.id)) {
                    notifyItemChanged(index)
                }
            }
        }

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): ChapterTrimViewHolder {
        val binding = ItemChapterTrimBinding.inflate(
            LayoutInflater.from(parent.context), parent, false
        )
        return ChapterTrimViewHolder(binding)
    }

    override fun onBindViewHolder(holder: ChapterTrimViewHolder, position: Int) {
        holder.bind(getItem(position), currentChapterId, selectedChapterIds, onChapterClick)
    }

    inner class ChapterTrimViewHolder(private val binding: ItemChapterTrimBinding) :
        RecyclerView.ViewHolder(binding.root) {

        fun bind(
            option: ChapterTrimOption,
            currentId: Long,
            selectedIds: Set<Long>,
            onClick: (ChapterTrimOption) -> Unit
        ) {
            binding.tvChapterTitle.text = "${option.index}. ${option.title}"

            val isSelected = selectedIds.contains(option.id)
            val isCurrent = option.id == currentId

            binding.cardContainer.isEnabled = option.status == TrimStatus.AVAILABLE

            when (option.status) {
                TrimStatus.PROCESSING -> {
                    binding.cardContainer.setCardBackgroundColor(COLOR_DISABLED)
                    binding.cardContainer.strokeColor = COLOR_BORDER_DEFAULT
                    binding.tvChapterTitle.setTextColor(COLOR_TEXT_DISABLED)
                    binding.tvStatus.text = "处理中"
                    binding.tvStatus.visibility = View.VISIBLE
                    binding.selectCircle.setCardBackgroundColor(Color.TRANSPARENT)
                    binding.selectCircle.strokeColor = COLOR_BORDER_DEFAULT
                    binding.selectDot.visibility = View.GONE
                }
                TrimStatus.TRIMMED -> {
                    binding.cardContainer.setCardBackgroundColor(COLOR_DISABLED)
                    binding.cardContainer.strokeColor = COLOR_BORDER_DEFAULT
                    binding.tvChapterTitle.setTextColor(COLOR_TEXT_DISABLED)
                    binding.tvStatus.text = "已精简"
                    binding.tvStatus.visibility = View.VISIBLE
                    binding.selectCircle.setCardBackgroundColor(Color.TRANSPARENT)
                    binding.selectCircle.strokeColor = COLOR_BORDER_DEFAULT
                    binding.selectDot.visibility = View.GONE
                }
                TrimStatus.AVAILABLE -> {
                    binding.tvStatus.visibility = View.GONE
                    if (isSelected) {
                        binding.cardContainer.setCardBackgroundColor(COLOR_SELECTED)
                        binding.cardContainer.strokeColor = COLOR_BORDER_SELECTED
                        binding.tvChapterTitle.setTextColor(COLOR_TEXT_SELECTED)
                        binding.selectCircle.setCardBackgroundColor(COLOR_BORDER_SELECTED)
                        binding.selectCircle.strokeColor = COLOR_BORDER_SELECTED
                        binding.selectDot.visibility = View.VISIBLE
                    } else {
                        binding.cardContainer.setCardBackgroundColor(COLOR_DEFAULT)
                        binding.cardContainer.strokeColor = COLOR_BORDER_DEFAULT
                        binding.tvChapterTitle.setTextColor(COLOR_TEXT_DEFAULT)
                        binding.selectCircle.setCardBackgroundColor(Color.TRANSPARENT)
                        binding.selectCircle.strokeColor = COLOR_BORDER_DEFAULT
                        binding.selectDot.visibility = View.GONE
                    }
                }
            }

            if (isCurrent) {
                binding.tvCurrent.visibility = View.VISIBLE
            } else {
                binding.tvCurrent.visibility = View.GONE
            }

            binding.root.setOnClickListener {
                if (option.status == TrimStatus.AVAILABLE) {
                    onClick(option)
                }
            }
        }
    }

    class ChapterTrimDiffCallback : DiffUtil.ItemCallback<ChapterTrimOption>() {
        override fun areItemsTheSame(oldItem: ChapterTrimOption, newItem: ChapterTrimOption): Boolean {
            return oldItem.id == newItem.id
        }

        override fun areContentsTheSame(oldItem: ChapterTrimOption, newItem: ChapterTrimOption): Boolean {
            return oldItem == newItem
        }
    }
}
