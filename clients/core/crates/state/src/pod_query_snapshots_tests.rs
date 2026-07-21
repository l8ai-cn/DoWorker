use agentcloud_types::proto_pod_v1::Pod;

use crate::pod_query_snapshots::PodQuerySnapshots;

fn pod(key: &str, status: &str) -> Pod {
    Pod {
        pod_key: key.to_string(),
        status: status.to_string(),
        ..Default::default()
    }
}

#[test]
fn keeps_each_query_result_in_its_own_snapshot() {
    let mut snapshots = PodQuerySnapshots::default();
    snapshots.replace("mobile:dev-org", vec![pod("completed-pod", "completed")]);
    snapshots.replace("sidebar:mine", vec![pod("running-pod", "running")]);

    assert_eq!(snapshots.get("mobile:dev-org")[0].pod_key, "completed-pod");
    assert_eq!(snapshots.get("sidebar:mine")[0].pod_key, "running-pod");
}
