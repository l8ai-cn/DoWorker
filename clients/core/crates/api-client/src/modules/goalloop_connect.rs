use crate::connect_call::connect_call;
use crate::error::ApiError;
use crate::ApiClient;
use agentcloud_types::proto_goalloop_v1 as lp;

impl ApiClient {
    pub async fn compile_loop_program_connect(
        &self,
        req: &lp::CompileLoopProgramRequest,
    ) -> Result<lp::CompileLoopProgramResponse, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/CompileLoopProgram",
            req,
        )
        .await
    }

    pub async fn generate_loop_program_connect(
        &self,
        req: &lp::GenerateLoopProgramRequest,
    ) -> Result<lp::CompileLoopProgramResponse, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/GenerateLoopProgram",
            req,
        )
        .await
    }

    pub async fn repair_loop_program_connect(
        &self,
        req: &lp::RepairLoopProgramRequest,
    ) -> Result<lp::RepairLoopProgramResponse, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/RepairLoopProgram",
            req,
        )
        .await
    }

    pub async fn run_loop_program_connect(
        &self,
        req: &lp::RunLoopProgramRequest,
    ) -> Result<lp::GoalLoop, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/RunLoopProgram",
            req,
        )
        .await
    }

    pub async fn list_worker_snapshots_connect(
        &self,
        req: &lp::ListWorkerSnapshotsRequest,
    ) -> Result<lp::ListWorkerSnapshotsResponse, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/ListWorkerSnapshots",
            req,
        )
        .await
    }

    pub async fn list_goal_loops_connect(
        &self,
        req: &lp::ListGoalLoopsRequest,
    ) -> Result<lp::ListGoalLoopsResponse, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/ListGoalLoops",
            req,
        )
        .await
    }

    pub async fn get_goal_loop_connect(
        &self,
        req: &lp::GetGoalLoopRequest,
    ) -> Result<lp::GoalLoop, ApiError> {
        connect_call(self, "/proto.goalloop.v1.GoalLoopService/GetGoalLoop", req).await
    }

    pub async fn create_goal_loop_connect(
        &self,
        req: &lp::CreateGoalLoopRequest,
    ) -> Result<lp::GoalLoop, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/CreateGoalLoop",
            req,
        )
        .await
    }

    pub async fn start_goal_loop_connect(
        &self,
        req: &lp::GoalLoopActionRequest,
    ) -> Result<lp::GoalLoop, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/StartGoalLoop",
            req,
        )
        .await
    }

    pub async fn verify_goal_loop_connect(
        &self,
        req: &lp::GoalLoopActionRequest,
    ) -> Result<lp::GoalLoop, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/VerifyGoalLoop",
            req,
        )
        .await
    }

    pub async fn cancel_goal_loop_connect(
        &self,
        req: &lp::GoalLoopActionRequest,
    ) -> Result<lp::GoalLoop, ApiError> {
        connect_call(
            self,
            "/proto.goalloop.v1.GoalLoopService/CancelGoalLoop",
            req,
        )
        .await
    }
}
