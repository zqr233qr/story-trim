package com.storytrim.app.ui.tasks.adapter

import android.view.LayoutInflater
import android.view.ViewGroup
import androidx.recyclerview.widget.DiffUtil
import androidx.recyclerview.widget.ListAdapter
import androidx.recyclerview.widget.RecyclerView
import com.storytrim.app.databinding.ItemTaskCardBinding
import com.storytrim.app.ui.tasks.TaskItemUi

class TaskItemAdapter : ListAdapter<TaskItemUi, TaskItemAdapter.TaskViewHolder>(DiffCallback()) {

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): TaskViewHolder {
        val binding = ItemTaskCardBinding.inflate(LayoutInflater.from(parent.context), parent, false)
        return TaskViewHolder(binding)
    }

    override fun onBindViewHolder(holder: TaskViewHolder, position: Int) {
        holder.bind(getItem(position))
    }

    class TaskViewHolder(private val binding: ItemTaskCardBinding) : RecyclerView.ViewHolder(binding.root) {
        fun bind(item: TaskItemUi) {
            binding.tvTitle.text = item.title
            binding.tvSubtitle.text = item.subtitle
            binding.tvStatus.text = item.status
            binding.progressBar.progress = item.progress.coerceIn(0, 100)
            binding.tvProgress.text = "${item.progress.coerceIn(0, 100)}%"
        }
    }

    private class DiffCallback : DiffUtil.ItemCallback<TaskItemUi>() {
        override fun areItemsTheSame(oldItem: TaskItemUi, newItem: TaskItemUi): Boolean {
            return oldItem.id == newItem.id
        }

        override fun areContentsTheSame(oldItem: TaskItemUi, newItem: TaskItemUi): Boolean {
            return oldItem == newItem
        }
    }
}
