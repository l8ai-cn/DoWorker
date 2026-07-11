use crate::ApiClient;
use crate::connect_call::connect_call;
use crate::error::ApiError;
use agentsmesh_types::proto_goalloop_v1 as lp;

impl ApiClient {
    pub async fn list_goal_loops_connect(
        &self,
        req: &lp::ListGoalLoopsRequest,
    ) -> Result<lp::ListGoalLoopsResponse, ApiError> {
        connect_call(self, "/proto.goalloop.v1.GoalLoopService/ListGoalLoops", req).await
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
        connect_call(self, "/proto.goalloop.v1.GoalLoopService/CreateGoalLoop", req).await
    }

    pub async fn start_goal_loop_connect(
        &self,
        req: &lp::GoalLoopActionRequest,
    ) -> Result<lp::GoalLoop, ApiError> {
        connect_call(self, "/proto.goalloop.v1.GoalLoopService/StartGoalLoop", req).await
    }

    pub async fn verify_goal_loop_connect(
        &self,
        req: &lp::GoalLoopActionRequest,
    ) -> Result<lp::GoalLoop, ApiError> {
        connect_call(self, "/proto.goalloop.v1.GoalLoopService/VerifyGoalLoop", req).await
    }

    pub async fn cancel_goal_loop_connect(
        &self,
        req: &lp::GoalLoopActionRequest,
    ) -> Result<lp::GoalLoop, ApiError> {
        connect_call(self, "/proto.goalloop.v1.GoalLoopService/CancelGoalLoop", req).await
    }
}
