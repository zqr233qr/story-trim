package com.storytrim.app.ui.reader.dialog

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.activityViewModels
import androidx.recyclerview.widget.LinearLayoutManager
import com.google.android.material.bottomsheet.BottomSheetBehavior
import com.google.android.material.bottomsheet.BottomSheetDialog
import com.google.android.material.bottomsheet.BottomSheetDialogFragment
import com.storytrim.app.databinding.FragmentChapterListBinding
import com.storytrim.app.ui.reader.ReaderViewModel
import com.storytrim.app.ui.reader.adapter.ChapterAdapter

class ChapterListDialogFragment : BottomSheetDialogFragment() {

    private var _binding: FragmentChapterListBinding? = null
    private val binding get() = _binding!!
    
    private val viewModel: ReaderViewModel by activityViewModels()
    private lateinit var adapter: ChapterAdapter

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentChapterListBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        
        // Make full height
        val dialog = dialog as? BottomSheetDialog
        dialog?.behavior?.state = BottomSheetBehavior.STATE_EXPANDED
        dialog?.behavior?.skipCollapsed = true
        
        // 80% height
        val displayMetrics = resources.displayMetrics
        val height = (displayMetrics.heightPixels * 0.85).toInt()
        binding.root.layoutParams.height = height

        setupRecyclerView()
        setupObservers()
    }

    private fun setupRecyclerView() {
        adapter = ChapterAdapter { index ->
            viewModel.jumpToChapter(index)
            dismiss()
        }
        binding.recyclerViewChapters.layoutManager = LinearLayoutManager(requireContext())
        binding.recyclerViewChapters.adapter = adapter
    }

    private fun setupObservers() {
        viewModel.chapters.observe(viewLifecycleOwner) { chapters ->
            binding.tvChapterCount.text = "${chapters.size} 章"
            adapter.submitList(chapters)
            
            // Scroll to current
            val current = viewModel.currentChapterIndex
            adapter.currentChapterIndex = current
            if (current >= 0 && current < chapters.size) {
                binding.recyclerViewChapters.scrollToPosition(current)
            }
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
