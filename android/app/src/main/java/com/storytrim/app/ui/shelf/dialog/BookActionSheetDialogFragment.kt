package com.storytrim.app.ui.shelf.dialog

import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import com.google.android.material.bottomsheet.BottomSheetDialogFragment
import com.storytrim.app.databinding.DialogBookActionSheetBinding

class BookActionSheetDialogFragment : BottomSheetDialogFragment() {

    enum class ActionType { SYNC, DOWNLOAD, DELETE }

    private var _binding: DialogBookActionSheetBinding? = null
    private val binding get() = _binding!!

    private var title: String = ""
    private var showSync: Boolean = true
    private var showDownload: Boolean = false
    private var listener: ((ActionType) -> Unit)? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        arguments?.let {
            title = it.getString(ARG_TITLE).orEmpty()
            showSync = it.getBoolean(ARG_SHOW_SYNC)
            showDownload = it.getBoolean(ARG_SHOW_DOWNLOAD)
        }
    }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = DialogBookActionSheetBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        binding.tvTitle.text = title
        binding.btnSync.visibility = if (showSync) View.VISIBLE else View.GONE
        binding.btnDownload.visibility = if (showDownload) View.VISIBLE else View.GONE

        binding.btnSync.setOnClickListener {
            listener?.invoke(ActionType.SYNC)
            dismiss()
        }
        binding.btnDownload.setOnClickListener {
            listener?.invoke(ActionType.DOWNLOAD)
            dismiss()
        }
        binding.btnDelete.setOnClickListener {
            listener?.invoke(ActionType.DELETE)
            dismiss()
        }
        binding.btnCancel.setOnClickListener { dismiss() }
        binding.handle.setOnClickListener { dismiss() }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }

    fun setActionListener(listener: (ActionType) -> Unit): BookActionSheetDialogFragment {
        this.listener = listener
        return this
    }

    companion object {
        private const val ARG_TITLE = "title"
        private const val ARG_SHOW_SYNC = "show_sync"
        private const val ARG_SHOW_DOWNLOAD = "show_download"

        fun newInstance(title: String, showSync: Boolean, showDownload: Boolean): BookActionSheetDialogFragment {
            val fragment = BookActionSheetDialogFragment()
            val args = Bundle()
            args.putString(ARG_TITLE, title)
            args.putBoolean(ARG_SHOW_SYNC, showSync)
            args.putBoolean(ARG_SHOW_DOWNLOAD, showDownload)
            fragment.arguments = args
            return fragment
        }
    }
}
