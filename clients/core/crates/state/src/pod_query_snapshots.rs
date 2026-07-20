use std::collections::HashMap;

use agentsmesh_types::proto_pod_v1::Pod;

#[derive(Default)]
pub struct PodQuerySnapshots {
    pods_by_query: HashMap<String, Vec<Pod>>,
}

impl PodQuerySnapshots {
    pub fn replace(&mut self, query_key: &str, pods: Vec<Pod>) {
        self.pods_by_query.insert(query_key.to_string(), pods);
    }

    pub fn get(&self, query_key: &str) -> &[Pod] {
        self.pods_by_query
            .get(query_key)
            .map(Vec::as_slice)
            .unwrap_or(&[])
    }
}
