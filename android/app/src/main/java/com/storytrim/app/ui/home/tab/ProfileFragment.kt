package com.storytrim.app.ui.home.tab

import android.content.Intent
import android.os.Bundle
import android.view.LayoutInflater
import android.view.View
import android.view.ViewGroup
import androidx.fragment.app.Fragment
import androidx.fragment.app.viewModels
import androidx.core.view.ViewCompat
import androidx.core.view.WindowInsetsCompat
import com.storytrim.app.databinding.FragmentProfileBinding
import com.storytrim.app.ui.login.LoginActivity
import com.storytrim.app.ui.common.ToastHelper
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class ProfileFragment : Fragment() {

    private var _binding: FragmentProfileBinding? = null
    private val binding get() = _binding!!
    private val viewModel: ProfileViewModel by viewModels()

    override fun onCreateView(
        inflater: LayoutInflater,
        container: ViewGroup?,
        savedInstanceState: Bundle?
    ): View {
        _binding = FragmentProfileBinding.inflate(inflater, container, false)
        return binding.root
    }

    override fun onViewCreated(view: View, savedInstanceState: Bundle?) {
        super.onViewCreated(view, savedInstanceState)

        ViewCompat.setOnApplyWindowInsetsListener(binding.root) { _, windowInsets ->
            val insets = windowInsets.getInsets(WindowInsetsCompat.Type.systemBars())
            val extraTop = (8 * resources.displayMetrics.density).toInt()
            binding.root.setPadding(
                binding.root.paddingLeft,
                insets.top + extraTop,
                binding.root.paddingRight,
                binding.root.paddingBottom
            )
            WindowInsetsCompat.CONSUMED
        }

        viewModel.username.observe(viewLifecycleOwner) { name ->
            val displayName = if (name.isNullOrBlank()) "未登录" else name
            binding.tvUsername.text = displayName
            binding.tvAvatar.text = displayName.first().uppercaseChar().toString()
            binding.btnLogin.visibility = if (name.isNullOrBlank()) View.VISIBLE else View.GONE
            binding.btnLogout.visibility = if (name.isNullOrBlank()) View.GONE else View.VISIBLE
            viewModel.loadPoints(name.isNotBlank())
        }

        viewModel.pointsBalance.observe(viewLifecycleOwner) { balance ->
            binding.tvPointsValue.text = "积分 $balance"
        }

        binding.btnLogin.setOnClickListener {
            startActivity(Intent(requireContext(), LoginActivity::class.java))
        }

        binding.btnLogout.setOnClickListener {
            viewModel.logout()
        }

        binding.itemPoints.setOnClickListener {
            startActivity(Intent(requireContext(), com.storytrim.app.ui.points.PointsActivity::class.java))
        }

        binding.itemTasks.setOnClickListener {
            startActivity(Intent(requireContext(), com.storytrim.app.ui.tasks.TaskCenterActivity::class.java))
        }

        viewModel.logoutResult.observe(viewLifecycleOwner) { result ->
            if (result == true) {
                ToastHelper.show(requireContext(), "已退出登录")
            }
        }
    }

    override fun onDestroyView() {
        super.onDestroyView()
        _binding = null
    }
}
