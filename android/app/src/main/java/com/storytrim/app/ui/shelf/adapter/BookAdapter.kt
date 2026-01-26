package com.storytrim.app.ui.shelf.adapter

import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.recyclerview.widget.DiffUtil
import androidx.recyclerview.widget.ListAdapter
import androidx.recyclerview.widget.RecyclerView
import com.storytrim.app.data.model.Book
import com.storytrim.app.databinding.ItemBookCardBinding

class BookAdapter(
    private val onBookClick: (Book) -> Unit,
    private val onBookLongClick: (Book) -> Unit
) : ListAdapter<Book, BookAdapter.BookViewHolder>(BookDiffCallback()) {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): BookViewHolder {
        val binding = ItemBookCardBinding.inflate(
            LayoutInflater.from(parent.context),
            parent,
            false
        )
        return BookViewHolder(binding, onBookClick, onBookLongClick)
    }

    override fun onBindViewHolder(holder: BookViewHolder, position: Int) {
        holder.bind(getItem(position))
    }

    class BookViewHolder(
        private val binding: ItemBookCardBinding,
        private val onBookClick: (Book) -> Unit,
        private val onBookLongClick: (Book) -> Unit
    ) : RecyclerView.ViewHolder(binding.root) {

        fun bind(book: Book) {
            binding.textViewTitle.text = book.title
            binding.textViewChapterCount.text = "${book.totalChapters} 章节"

            binding.root.setOnClickListener {
                onBookClick(book)
            }

            binding.root.setOnLongClickListener {
                onBookLongClick(book)
                true
            }

            // Sync State Logic:
            // 0: Local only -> Show "本地"
            // 1: Synced -> Show "本地" + "云端"
            // 2: Cloud only -> Show "云端"
            
            when (book.syncState) {
                0 -> {
                    binding.tvTagLocal.visibility = View.VISIBLE
                    binding.tvTagCloud.visibility = View.GONE
                }
                1 -> {
                    binding.tvTagLocal.visibility = View.VISIBLE
                    binding.tvTagCloud.visibility = View.VISIBLE
                }
                2 -> {
                    binding.tvTagLocal.visibility = View.GONE
                    binding.tvTagCloud.visibility = View.VISIBLE
                }
                else -> {
                    binding.tvTagLocal.visibility = View.VISIBLE
                    binding.tvTagCloud.visibility = View.GONE
                }
            }
        }
    }

    private class BookDiffCallback : DiffUtil.ItemCallback<Book>() {
        override fun areItemsTheSame(oldItem: Book, newItem: Book): Boolean {
            return oldItem.id == newItem.id
        }

        override fun areContentsTheSame(oldItem: Book, newItem: Book): Boolean {
            return oldItem == newItem
        }
    }
}
