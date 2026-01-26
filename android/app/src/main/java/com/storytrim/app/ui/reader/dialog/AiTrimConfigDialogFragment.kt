package com.storytrim.app.ui.reader.dialog

import android.graphics.Color
import android.graphics.drawable.ColorDrawable
import android.os.Bundle
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.lifecycle.lifecycleScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import androidx.fragment.app.DialogFragment
import androidx.fragment.app.activityViewModels
import androidx.recyclerview.widget.LinearLayoutManager
import com.storytrim.app.databinding.DialogAiTrimConfigBinding
import com.storytrim.app.ui.reader.ReaderViewModel
import com.storytrim.app.ui.reader.adapter.AiTrimOptionAdapter

class AiTrimConfigDialogFragment : DialogFragment() {

    private var _binding: DialogAiTrimConfigBinding? = null
    private val binding get() = _binding!!

    private val viewModel: ReaderViewModel by activityViewModels()
    private var adapter: AiTrimOptionAdapter? = null
    private var selectedPromptId: Int = 0
    private var trimmedPromptIds: Set<Int> = emptySet()

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = DialogAiTrimConfigBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        binding.btnLater.setOnClickListener { dismiss() }
        binding.btnStart.setOnClickListener {
            if (selectedPromptId > 0) {
                viewModel.switchMode(selectedPromptId)
            }
            dismiss()
        }

        viewModel.loadPrompts()
        viewModel.prompts.observe(viewLifecycleOwner) { prompts ->
            if (prompts.isEmpty()) return@observe
            val initialId = when {
                viewModel.userPreferredModeId > 0 -> viewModel.userPreferredModeId
                selectedPromptId > 0 -> selectedPromptId
                else -> prompts.first().id
            }
            selectedPromptId = initialId
            binding.recyclerOptions.layoutManager = LinearLayoutManager(requireContext())
            val adapter = this.adapter ?: AiTrimOptionAdapter(prompts, initialId, trimmedPromptIds) { modeId ->
                selectedPromptId = modeId
                this.adapter?.updateSelectedId(modeId)
                updateActionButton()
            }.also { this.adapter = it }
            adapter.updateSelectedId(initialId)
            adapter.updateTrimmedIds(trimmedPromptIds)
            binding.recyclerOptions.adapter = adapter

            val book = viewModel.book.value
            val chapter = viewModel.chapters.value?.getOrNull(viewModel.currentChapterIndex)
            if (book != null && chapter != null) {
                binding.tvSubtitle.text = "《${book.title}》第${chapter.index}章 ${chapter.title}"
                loadTrimStatus(book.syncState, chapter, book.bookMd5)
            }
        }
    }

    override fun onStart() {
        super.onStart()
        dialog?.window?.let { window ->
            window.setLayout(ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.WRAP_CONTENT)
            window.setGravity(Gravity.BOTTOM)
            window.setBackgroundDrawable(ColorDrawable(Color.TRANSPARENT))
        }
    }

    private fun loadTrimStatus(syncState: Int, chapter: com.storytrim.app.data.model.Chapter, bookMd5: String?) {
        lifecycleScope.launch(Dispatchers.IO) {
            val ids = if (syncState == 0 && chapter.md5.isNotBlank()) {
                viewModel.bookRepository.getChapterTrimStatusByMd5(chapter.md5)
            } else {
                viewModel.bookRepository.getChapterTrimStatusById(chapter, bookMd5)
            }
            trimmedPromptIds = ids.toSet()
            withContext(Dispatchers.Main) {
                adapter?.updateTrimmedIds(trimmedPromptIds)
                updateActionButton()
            }
        }
    }

    private fun updateActionButton() {
        val isTrimmed = trimmedPromptIds.contains(selectedPromptId)
        binding.btnStart.text = if (isTrimmed) "开始阅读" else "开始精简"
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }

    companion object {
        const val TAG = "AiTrimConfigDialog"
    }
}
