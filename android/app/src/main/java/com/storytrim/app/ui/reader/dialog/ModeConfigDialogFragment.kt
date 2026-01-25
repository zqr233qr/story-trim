package com.storytrim.app.ui.reader.dialog

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.activityViewModels
import androidx.recyclerview.widget.LinearLayoutManager
import com.google.android.material.bottomsheet.BottomSheetDialogFragment
import com.storytrim.app.databinding.DialogModeConfigBinding
import com.storytrim.app.ui.reader.ReaderViewModel
import com.storytrim.app.ui.reader.adapter.ModeAdapter

class ModeConfigDialogFragment : BottomSheetDialogFragment() {

    private var _binding: DialogModeConfigBinding? = null
    private val binding get() = _binding!!
    
    private val viewModel: ReaderViewModel by activityViewModels()

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = DialogModeConfigBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        
        // Load prompts if needed
        viewModel.loadPrompts()
        
        viewModel.prompts.observe(viewLifecycleOwner) { prompts ->
            val displayPrompts = listOf(
                com.storytrim.app.data.model.Prompt(
                    id = 0,
                    name = "原文",
                    description = "",
                    isDefault = false
                )
            ) + prompts
            val adapter = ModeAdapter(displayPrompts, viewModel.userPreferredModeId) { modeId ->
                viewModel.updateUserPreferredMode(modeId)
                dismiss()
            }
            binding.recyclerViewModes.layoutManager = LinearLayoutManager(requireContext())
            binding.recyclerViewModes.adapter = adapter
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
