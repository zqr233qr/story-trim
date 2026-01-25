package com.storytrim.app.ui.reader.dialog

import android.graphics.Color
import android.graphics.drawable.ColorDrawable
import android.os.Bundle
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.fragment.app.DialogFragment
import androidx.fragment.app.activityViewModels
import androidx.recyclerview.widget.GridLayoutManager
import com.storytrim.app.R
import com.storytrim.app.data.dto.ChapterTrimOption
import com.storytrim.app.data.dto.TrimStatus
import com.storytrim.app.databinding.FragmentChapterTrimBinding
import com.storytrim.app.ui.reader.ReaderViewModel
import com.storytrim.app.ui.reader.adapter.ChapterTrimAdapter
import com.storytrim.app.ui.reader.adapter.ModeAdapter
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class ChapterTrimDialogFragment : DialogFragment() {

    private var _binding: FragmentChapterTrimBinding? = null
    private val binding get() = _binding!!

    private val viewModel: ReaderViewModel by activityViewModels()
    private lateinit var chapterAdapter: ChapterTrimAdapter
    private lateinit var modeAdapter: ModeAdapter

    private var selectedPromptId = 0
    private var selectedChapterIds = mutableSetOf<Long>()
    private var pointsBalance = 0
    private var chapterOptions: List<ChapterTrimOption> = emptyList()

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentChapterTrimBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        setupRecyclerViews()
        setupClickListeners()
        loadPromptsAndBalance()
    }

    override fun onStart() {
        super.onStart()
        dialog?.window?.let { window ->
            val displayMetrics = resources.displayMetrics
            val height = (displayMetrics.heightPixels * 0.85).toInt()
            window.setLayout(ViewGroup.LayoutParams.MATCH_PARENT, height)
            window.setGravity(Gravity.BOTTOM)
            window.setBackgroundDrawable(ColorDrawable(Color.TRANSPARENT))
        }
    }

    private fun setupRecyclerViews() {
        val currentChapterId = viewModel.chapters.value?.getOrNull(viewModel.currentChapterIndex)?.id ?: 0L

        // Chapter Adapter
        chapterAdapter = ChapterTrimAdapter(
            onChapterClick = { option ->
                toggleChapterSelection(option)
            },
            currentChapterId = currentChapterId
        )
        binding.recyclerViewChapters.layoutManager = GridLayoutManager(requireContext(), 2)
        binding.recyclerViewChapters.isNestedScrollingEnabled = true
        binding.recyclerViewChapters.setHasFixedSize(true)
        binding.recyclerViewChapters.adapter = chapterAdapter
    }

    private fun setupClickListeners() {
        binding.btnQuickSelect3.setOnClickListener { quickSelect(3) }
        binding.btnQuickSelect5.setOnClickListener { quickSelect(5) }
        binding.btnQuickSelect10.setOnClickListener { quickSelect(10) }
        binding.btnQuickSelect20.setOnClickListener { quickSelect(20) }

        binding.btnConfirm.setOnClickListener {
            if (selectedChapterIds.isEmpty()) {
                Toast.makeText(requireContext(), "请选择章节", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            if (selectedChapterIds.size > pointsBalance) {
                Toast.makeText(requireContext(), "积分不足", Toast.LENGTH_SHORT).show()
                return@setOnClickListener
            }
            submitChapterTrimTask()
        }
    }

    private fun loadPromptsAndBalance() {
        val cachedPrompts = viewModel.prompts.value
        android.util.Log.d("ChapterTrimDialog", "loadPromptsAndBalance: prompts=${cachedPrompts?.size}")

        if (cachedPrompts == null) {
            android.util.Log.d("ChapterTrimDialog", "Loading prompts...")
            viewModel.loadPrompts()
        }

        viewModel.prompts.observe(viewLifecycleOwner) { prompts ->
            android.util.Log.d("ChapterTrimDialog", "Prompts received: ${prompts.size}")

            if (prompts.isNotEmpty()) {
                val defaultPrompt = if (viewModel.userPreferredModeId > 0) {
                    prompts.find { it.id == viewModel.userPreferredModeId }
                } else {
                    prompts.find { it.isDefault }
                } ?: prompts.first()
                selectedPromptId = defaultPrompt.id
            } else {
                selectedPromptId = 0
            }

            modeAdapter = ModeAdapter(prompts, selectedPromptId) { promptId ->
                selectedPromptId = promptId
                modeAdapter.updateSelectedId(promptId)
                android.util.Log.d("ChapterTrimDialog", "Mode changed to: $promptId")
                selectedChapterIds.clear()
                chapterAdapter.selectedChapterIds = selectedChapterIds
                updateConfirmButton()
                loadChapterTrimStatus()
            }
            binding.recyclerViewModes.adapter = modeAdapter

            loadPointsBalance()
            if (selectedPromptId > 0) {
                loadChapterTrimStatus()
            }
        }
    }

    private fun loadPointsBalance() {
        CoroutineScope(Dispatchers.IO).launch {
            viewModel.bookRepository.getPointsBalance()
                .onSuccess { balance ->
                    withContext(Dispatchers.Main) {
                        pointsBalance = balance
                        binding.tvPointsBalance.text = "积分 $balance"
                        updateConfirmButton()
                    }
                }
                .onFailure { e ->
                    withContext(Dispatchers.Main) {
                        Toast.makeText(requireContext(), "获取积分失败: ${e.message}", Toast.LENGTH_SHORT).show()
                    }
                }
        }
    }

    private fun loadChapterTrimStatus() {
        val book = viewModel.book.value ?: return
        val chapters = viewModel.chapters.value ?: return

        if (selectedPromptId == 0) return

        CoroutineScope(Dispatchers.IO).launch {
            viewModel.bookRepository.getChapterTrimStatus(book.cloudId, selectedPromptId, chapters)
                .onSuccess { options ->
                    withContext(Dispatchers.Main) {
                        chapterOptions = options
                        chapterAdapter.submitList(options)
                        updateConfirmButton()
                        scrollToCurrentChapter()
                    }
                }
                .onFailure { e ->
                    withContext(Dispatchers.Main) {
                        Toast.makeText(requireContext(), "获取精简状态失败: ${e.message}", Toast.LENGTH_SHORT).show()
                    }
                }
        }
    }

    private fun toggleChapterSelection(option: ChapterTrimOption) {
        if (option.status != TrimStatus.AVAILABLE) return

        if (selectedChapterIds.contains(option.id)) {
            selectedChapterIds.remove(option.id)
        } else {
            selectedChapterIds.add(option.id)
        }

        chapterAdapter.selectedChapterIds = selectedChapterIds.toSet()
        updateConfirmButton()
    }

    private fun quickSelect(count: Int) {
        val currentChapterIndex = viewModel.currentChapterIndex
        val startIndex = currentChapterIndex + 1

        if (startIndex >= chapterOptions.size) {
            Toast.makeText(requireContext(), "没有后续章节", Toast.LENGTH_SHORT).show()
            return
        }

        val candidates = chapterOptions
            .slice(startIndex until chapterOptions.size)
            .filter { it.status == TrimStatus.AVAILABLE }

        val toSelect = candidates.take(count).map { it.id }
        selectedChapterIds.clear()
        selectedChapterIds.addAll(toSelect)

        chapterAdapter.selectedChapterIds = selectedChapterIds.toSet()
        updateConfirmButton()
    }

    private fun scrollToCurrentChapter() {
        val currentChapterId = viewModel.chapters.value?.getOrNull(viewModel.currentChapterIndex)?.id ?: 0L
        val position = chapterOptions.indexOfFirst { it.id == currentChapterId }
        if (position >= 0) {
            binding.recyclerViewChapters.scrollToPosition(position)
        }
    }

    private fun updateConfirmButton() {
        val selectedCount = selectedChapterIds.size
        binding.tvSelectedInfo.text = "已选 $selectedCount 章 · 预计消耗 $selectedCount 积分"

        val hasEnoughBalance = selectedCount > 0 && selectedCount <= pointsBalance
        binding.btnConfirm.isEnabled = hasEnoughBalance
        if (hasEnoughBalance) {
            binding.btnConfirm.setBackgroundResource(R.drawable.bg_primary_button)
            binding.btnConfirm.setTextColor(android.graphics.Color.parseColor("#FFFFFF"))
        } else {
            binding.btnConfirm.setBackgroundResource(R.drawable.bg_primary_button_disabled)
            binding.btnConfirm.setTextColor(android.graphics.Color.parseColor("#A8A29E"))
        }
    }

    private fun submitChapterTrimTask() {
        val book = viewModel.book.value ?: return

        CoroutineScope(Dispatchers.IO).launch {
            viewModel.bookRepository.startChapterTrimTask(book.cloudId, selectedPromptId, selectedChapterIds.toList())
                .onSuccess {
                    withContext(Dispatchers.Main) {
                        Toast.makeText(requireContext(), "任务已创建，可在书架查看进度", Toast.LENGTH_SHORT).show()
                        dismiss()
                    }
                }
                .onFailure { e ->
                    withContext(Dispatchers.Main) {
                        val errorMsg = when (e.message) {
                            "积分不足" -> "积分不足"
                            "章节已精简或处理中" -> "章节已精简或处理中"
                            else -> "任务创建失败: ${e.message}"
                        }
                        Toast.makeText(requireContext(), errorMsg, Toast.LENGTH_SHORT).show()
                    }
                }
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
