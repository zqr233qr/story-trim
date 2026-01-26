package com.storytrim.app.ui.reader.dialog

import android.graphics.Color
import android.graphics.drawable.ColorDrawable
import android.os.Bundle
import android.view.Gravity
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.activity.result.contract.ActivityResultContracts
import androidx.fragment.app.DialogFragment
import androidx.fragment.app.activityViewModels
import androidx.recyclerview.widget.GridLayoutManager
import com.storytrim.app.databinding.DialogReaderSettingsBinding
import com.storytrim.app.ui.reader.ReaderActivity
import com.storytrim.app.ui.reader.ReaderViewModel
import com.storytrim.app.ui.reader.adapter.ModeGridAdapter
import com.storytrim.app.ui.common.ToastHelper

class ReaderSettingsDialogFragment : DialogFragment() {

    private var _binding: DialogReaderSettingsBinding? = null
    private val binding get() = _binding!!

    private val viewModel: ReaderViewModel by activityViewModels()
    private var modeAdapter: ModeGridAdapter? = null
    private var readerActivity: ReaderActivity? = null

    private val pickBackground = registerForActivityResult(ActivityResultContracts.GetContent()) { uri ->
        val activity = readerActivity ?: return@registerForActivityResult
        if (uri == null) return@registerForActivityResult
        val target = java.io.File(activity.filesDir, "reader_background.jpg")
        try {
            activity.contentResolver.openInputStream(uri)?.use { input ->
                target.outputStream().use { output ->
                    input.copyTo(output)
                }
            }
            activity.updateBackgroundImage(target.absolutePath)
            updateBackgroundStatus(target.absolutePath)
        } catch (e: Exception) {
            ToastHelper.show(requireContext(), "背景图设置失败")
        }
    }

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = DialogReaderSettingsBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        val host = activity as? ReaderActivity ?: return
        readerActivity = host
        setupFontSize(host)
        setupReadingMode(host)
        setupPageMode(host)
        setupBackground(host)
        setupPreferredMode()

        viewModel.loadPrompts()
    }

    override fun onStart() {
        super.onStart()
        dialog?.window?.let { window ->
            val displayMetrics = resources.displayMetrics
            val maxHeight = (displayMetrics.heightPixels * 0.7).toInt()
            window.setLayout(ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.WRAP_CONTENT)
            window.setGravity(Gravity.BOTTOM)
            window.setBackgroundDrawable(ColorDrawable(Color.TRANSPARENT))

            binding.settingsScrollView.post {
                val layoutParams = binding.settingsScrollView.layoutParams
                layoutParams.height = binding.settingsScrollView.height.coerceAtMost(maxHeight)
                binding.settingsScrollView.layoutParams = layoutParams
            }
        }
        viewModel.setToastSuppressed(true)
    }

    private fun setupFontSize(activity: ReaderActivity) {
        val current = activity.getCurrentFontSizeSp().toInt().coerceIn(14, 26)
        val progress = (current - 14).coerceIn(0, 12)
        binding.seekFontSize.progress = progress
        binding.seekFontSize.setOnSeekBarChangeListener(object : android.widget.SeekBar.OnSeekBarChangeListener {
            override fun onProgressChanged(seekBar: android.widget.SeekBar?, progress: Int, fromUser: Boolean) {
                if (!fromUser) return
                val size = (14 + progress).toFloat()
                activity.updateFontSize(size)
            }

            override fun onStartTrackingTouch(seekBar: android.widget.SeekBar?) = Unit
            override fun onStopTrackingTouch(seekBar: android.widget.SeekBar?) = Unit
        })
    }

    private fun setupReadingMode(activity: ReaderActivity) {
        when (activity.getCurrentReadingMode()) {
            "dark" -> binding.toggleReadingMode.check(binding.btnReadingDark.id)
            "sepia" -> binding.toggleReadingMode.check(binding.btnReadingSepia.id)
            else -> binding.toggleReadingMode.check(binding.btnReadingLight.id)
        }

        binding.toggleReadingMode.addOnButtonCheckedListener { _, checkedId, isChecked ->
            if (!isChecked) return@addOnButtonCheckedListener
            val mode = when (checkedId) {
                binding.btnReadingDark.id -> "dark"
                binding.btnReadingSepia.id -> "sepia"
                else -> "light"
            }
            activity.updateReadingMode(mode)
        }
    }

    private fun setupPageMode(activity: ReaderActivity) {
        when (activity.getCurrentPageMode()) {
            "scroll" -> binding.togglePageMode.check(binding.btnPageScroll.id)
            else -> binding.togglePageMode.check(binding.btnPageClick.id)
        }

        binding.togglePageMode.addOnButtonCheckedListener { _, checkedId, isChecked ->
            if (!isChecked) return@addOnButtonCheckedListener
            val mode = if (checkedId == binding.btnPageScroll.id) "scroll" else "click"
            activity.updatePageMode(mode)
        }
    }

    private fun setupBackground(activity: ReaderActivity) {
        updateBackgroundStatus(activity.getCurrentBackgroundPath())
        binding.btnSelectBg.setOnClickListener {
            pickBackground.launch("image/*")
        }
        binding.btnClearBg.setOnClickListener {
            val target = java.io.File(activity.filesDir, "reader_background.jpg")
            if (target.exists()) {
                target.delete()
            }
            activity.updateBackgroundImage(null)
            updateBackgroundStatus(null)
        }
    }

    private fun updateBackgroundStatus(path: String?) {
        val hasBackground = !path.isNullOrBlank()
        binding.tvBgStatus.text = if (hasBackground) "已设置" else "未设置"
        binding.btnClearBg.visibility = if (hasBackground) View.VISIBLE else View.GONE
        binding.btnSelectBg.text = if (hasBackground) "更换" else "选择"
    }

    private fun setupPreferredMode() {
        viewModel.prompts.observe(viewLifecycleOwner) { prompts ->
            val selectedId = viewModel.userPreferredModeId
            val adapter = modeAdapter ?: ModeGridAdapter(prompts, selectedId) { modeId ->
                viewModel.updateUserPreferredMode(modeId)
                modeAdapter?.updateSelectedId(modeId)
            }.also { modeAdapter = it }

            adapter.updateSelectedId(selectedId)
            binding.recyclerPreferredModes.layoutManager = GridLayoutManager(requireContext(), 4)
            binding.recyclerPreferredModes.adapter = adapter
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        viewModel.setToastSuppressed(false)
        _binding = null
    }

    companion object {
        const val TAG = "ReaderSettingsDialog"

        fun newInstance(): ReaderSettingsDialogFragment {
            return ReaderSettingsDialogFragment()
        }
    }
}
