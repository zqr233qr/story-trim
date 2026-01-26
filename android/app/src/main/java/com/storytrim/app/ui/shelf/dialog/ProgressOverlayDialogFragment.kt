package com.storytrim.app.ui.shelf.dialog

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.DialogFragment
import com.storytrim.app.databinding.DialogProgressOverlayBinding

class ProgressOverlayDialogFragment : DialogFragment() {

    private var _binding: DialogProgressOverlayBinding? = null
    private val binding get() = _binding!!

    private var title: String = ""
    private var progress: Int = 0

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            title = it.getString(ARG_TITLE).orEmpty()
            progress = it.getInt(ARG_PROGRESS)
        }
        isCancelable = false
    }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = DialogProgressOverlayBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)
        updateProgress(progress, title)
    }

    fun updateProgress(value: Int, titleOverride: String? = null) {
        val clamped = value.coerceIn(0, 100)
        if (titleOverride != null) {
            title = titleOverride
        }
        binding.tvProgressTitle.text = title
        binding.progressBar.progress = clamped
        binding.tvProgressValue.text = "$clamped%"
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }

    companion object {
        private const val ARG_TITLE = "title"
        private const val ARG_PROGRESS = "progress"

        fun newInstance(title: String, progress: Int = 0): ProgressOverlayDialogFragment {
            val fragment = ProgressOverlayDialogFragment()
            val args = Bundle()
            args.putString(ARG_TITLE, title)
            args.putInt(ARG_PROGRESS, progress)
            fragment.arguments = args
            return fragment
        }
    }
}
