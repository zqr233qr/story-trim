package com.storytrim.app.ui.reader.adapter

import android.graphics.Color
import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.DiffUtil
import androidx.recyclerview.widget.ListAdapter
import androidx.recyclerview.widget.RecyclerView
import com.storytrim.app.data.model.Chapter
import com.storytrim.app.databinding.ItemChapterBinding

class ChapterAdapter(
    private val onChapterClick: (Int) -> Unit
) : ListAdapter<Chapter, ChapterAdapter.ChapterViewHolder>(ChapterDiffCallback()) {

    var currentChapterIndex: Int = -1

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): ChapterViewHolder {
        val binding = ItemChapterBinding.inflate(
            LayoutInflater.from(parent.context), parent, false
        )
        return ChapterViewHolder(binding)
    }

    override fun onBindViewHolder(holder: ChapterViewHolder, position: Int) {
        holder.bind(getItem(position), position == currentChapterIndex, position)
    }

    inner class ChapterViewHolder(private val binding: ItemChapterBinding) : 
        RecyclerView.ViewHolder(binding.root) {
        
        fun bind(chapter: Chapter, isSelected: Boolean, position: Int) {
            binding.tvChapterTitle.text = chapter.title
            
            if (isSelected) {
                binding.tvChapterTitle.setTextColor(Color.parseColor("#0D9488")) // Teal-600
                binding.root.setBackgroundColor(Color.parseColor("#F0FDFA")) // Teal-50
            } else {
                binding.tvChapterTitle.setTextColor(Color.parseColor("#1C1917")) // Stone-900
                binding.root.setBackgroundColor(Color.TRANSPARENT)
            }

            binding.root.setOnClickListener {
                onChapterClick(position)
            }
        }
    }

    class ChapterDiffCallback : DiffUtil.ItemCallback<Chapter>() {
        override fun areItemsTheSame(oldItem: Chapter, newItem: Chapter): Boolean {
            return oldItem.id == newItem.id
        }

        override fun areContentsTheSame(oldItem: Chapter, newItem: Chapter): Boolean {
            return oldItem == newItem
        }
    }
}
