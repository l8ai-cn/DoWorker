import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { PromoCodeType } from "@/lib/api/admin";
import type { PromoCodeFormData } from "./promo-code-form";

interface PromoCodeBasicFieldsProps {
  formData: PromoCodeFormData;
  onFormChange: (data: PromoCodeFormData) => void;
}

export function PromoCodeBasicFields({
  formData,
  onFormChange,
}: PromoCodeBasicFieldsProps) {
  return (
    <>
      <div className="space-y-2">
        <Label htmlFor="code">代码 *</Label>
        <Input
          id="code"
          placeholder="例如：SUMMER2026"
          value={formData.code}
          onChange={(e) =>
            onFormChange({ ...formData, code: e.target.value.toUpperCase() })
          }
          required
          minLength={4}
          maxLength={50}
          className="font-mono uppercase"
        />
        <p className="text-xs text-muted-foreground">
          4-50 个字符，将自动转换为大写
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="name">名称 *</Label>
        <Input
          id="name"
          placeholder="例如：夏季订阅活动"
          value={formData.name}
          onChange={(e) => onFormChange({ ...formData, name: e.target.value })}
          required
          maxLength={100}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="description">描述</Label>
        <Textarea
          id="description"
          placeholder="可选描述..."
          value={formData.description}
          onChange={(e) =>
            onFormChange({ ...formData, description: e.target.value })
          }
          rows={3}
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label>类型 *</Label>
          <Select
            value={formData.type}
            onValueChange={(value) =>
              onFormChange({ ...formData, type: value as PromoCodeType })
            }
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="media">媒体</SelectItem>
              <SelectItem value="partner">合作伙伴</SelectItem>
              <SelectItem value="campaign">活动</SelectItem>
              <SelectItem value="internal">内部</SelectItem>
              <SelectItem value="referral">推荐</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>套餐 *</Label>
          <Select
            value={formData.plan_name}
            onValueChange={(value) =>
              onFormChange({ ...formData, plan_name: value })
            }
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="pro">Pro</SelectItem>
              <SelectItem value="enterprise">Enterprise</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>
    </>
  );
}
