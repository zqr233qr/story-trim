package com.storytrim.app.ui.reader.dialog

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import android.widget.Toast
import androidx.fragment.app.activityViewModels
import com.google.android.material.bottomsheet.BottomSheetBehavior
import com.google.android.material.bottomsheet.BottomSheetDialog
import com.google.android.material.bottomsheet.BottomSheetDialogFragment
import com.storytrim.app.databinding.FragmentGenerationTerminalBinding
import com.storytrim.app.ui.reader.ReaderViewModel
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

class GenerationTerminalDialogFragment : BottomSheetDialogFragment() {

    private var _binding: FragmentGenerationTerminalBinding? = null
    private val binding get() = _binding!!

    private val viewModel: ReaderViewModel by activityViewModels()

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentGenerationTerminalBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        // 设置60%高度
        val dialog = dialog as? BottomSheetDialog
        dialog?.behavior?.state = BottomSheetBehavior.STATE_EXPANDED
        dialog?.behavior?.skipCollapsed = true

        val displayMetrics = resources.displayMetrics
        val height = (displayMetrics.heightPixels * 0.6).toInt()
        binding.root.layoutParams.height = height

        binding.btnCancel.setOnClickListener {
            dismiss()
        }

        viewModel.generationStream.observe(viewLifecycleOwner) { text ->
            binding.tvContent.text = text
            binding.scrollViewTerminal.post {
                binding.scrollViewTerminal.fullScroll(View.FOCUS_DOWN)
            }
        }

        viewModel.isGenerating.observe(viewLifecycleOwner) { isGenerating ->
            if (!isGenerating) {
                // 自动关闭延迟800ms
                CoroutineScope(Dispatchers.Main).launch {
                    delay(800)
                    if (isAdded) {
                        dismiss()
                        Toast.makeText(requireContext(), "精简完成", Toast.LENGTH_SHORT).show()
                    }
                }
            }
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }

    companion object {
        const val TAG = "GenerationTerminalDialog"
    }
}
